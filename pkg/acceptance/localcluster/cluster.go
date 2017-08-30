// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package localcluster

import (
	"bytes"
	gosql "database/sql"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/pkg/errors"
	// Import postgres driver.
	_ "github.com/lib/pq"
	"golang.org/x/net/context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/internal/client"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

func repoRoot() string {
	root, err := build.Import("github.com/cockroachdb/cockroach", "", build.FindOnly)
	if err != nil {
		panic(fmt.Sprintf("must run from within the cockroach repository: %s", err))
	}
	return root.Dir
}

const listeningURLFile = "cockroachdb-url"

// IsUnavailableError returns true iff the error corresponds to a GRPC
// connection unavailable error.
func IsUnavailableError(err error) bool {
	return strings.Contains(err.Error(), "grpc: the connection is unavailable")
}

// A ClusterConfig holds the configuration for a Cluster.
type ClusterConfig struct {
	Ephemeral   bool               // when true, wipe DataDir on Close()
	Binary      string             // path to cockroach, defaults go <cockroach_repo>/cockroach
	AllNodeArgs []string           // args to pass to ./cockroach on all nodes
	NumNodes    int                // number of nodes in the cluster
	DataDir     string             // node i will use storage DataDir/<i>
	LogDir      string             // when empty, node i defaults to DataDir/<i>/logs
	PerNodeCfg  map[int]NodeConfig // optional map of nodeIndex -> configuration
	DB          string             // database to configure DB connection for
	NumWorkers  int                // SetMaxOpenConns to use for DB connection
}

// NodeConfig is a configuration for a node in a Cluster. Options with the zero
// value are typically populated from the corresponding Cluster's ClusterConfig.
type NodeConfig struct {
	Binary            string   // when specified, overrides the node's binary
	DataDir           string   // when specified, overrides the node's data dir
	LogDir            string   // when specified, overrides the node's log dir
	Addr              string   // listening host, defaults to 127.0.0.1
	ExtraArgs         []string // extra arguments for ./cockroach start
	ExtraEnv          []string // environment variables in format key=value
	RPCPort, HTTPPort int      // zero for auto-assign
	DB                string   // see ClusterConfig
	NumWorkers        int      // see ClusterConfig
}

// MakePerNodeFixedPortsCfg makes a PerNodeCfg map of the given number of nodes
// with odd ports starting at 26257 for the RPC endpoint, and even points for
// the ui.
func MakePerNodeFixedPortsCfg(numNodes int) map[int]NodeConfig {
	perNodeCfg := make(map[int]NodeConfig)

	for i := 0; i < numNodes; i++ {
		perNodeCfg[i] = NodeConfig{
			RPCPort:  26257 + 2*i,
			HTTPPort: 26258 + 2*i,
		}
	}

	return perNodeCfg
}

// Cluster holds the state for a local cluster, providing methods for common
// operations, access to the underlying nodes and per-node KV and SQL clients.
type Cluster struct {
	cfg     ClusterConfig
	seq     *seqGen
	Nodes   []*Node
	stopper *stop.Stopper
	started time.Time
}

type seqGen int32

func (s *seqGen) Next() int32 {
	return atomic.AddInt32((*int32)(s), 1)
}

// New creates a Cluster with the given configuration.
func New(cfg ClusterConfig) *Cluster {
	if cfg.Binary == "" {
		cfg.Binary = filepath.Join(repoRoot(), "cockroach")
	}
	return &Cluster{
		cfg:     cfg,
		seq:     new(seqGen),
		stopper: stop.NewStopper(),
	}
}

// Start starts a cluster. The numWorkers parameter controls the SQL connection
// settings to avoid unnecessary connection creation. The allNodeArgs parameter
// can be used to pass extra arguments to every node. The perNodeArgs parameter
// can be used to pass extra arguments to an individual node. If not nil, its
// size must equal the number of nodes.
func (c *Cluster) Start(ctx context.Context) {
	c.started = timeutil.Now()

	chs := make([]<-chan error, c.cfg.NumNodes)
	for i := 0; i < c.cfg.NumNodes; i++ {
		cfg := c.cfg.PerNodeCfg[i] // zero value is ok
		if cfg.Binary == "" {
			cfg.Binary = c.cfg.Binary
		}
		if cfg.DataDir == "" {
			cfg.DataDir = filepath.Join(c.cfg.DataDir, fmt.Sprintf("%d", i+1))
		}
		if cfg.LogDir == "" && c.cfg.LogDir != "" {
			cfg.LogDir = filepath.Join(c.cfg.LogDir, fmt.Sprintf("%d", i+1))
		}
		if cfg.Addr == "" {
			cfg.Addr = "127.0.0.1"
		}
		if cfg.DB == "" {
			cfg.DB = c.cfg.DB
		}
		if cfg.NumWorkers == 0 {
			cfg.NumWorkers = c.cfg.NumWorkers
		}
		cfg.ExtraArgs = append(append([]string(nil), c.cfg.AllNodeArgs...), cfg.ExtraArgs...)
		var node *Node
		node, chs[i] = c.makeNode(ctx, i, cfg)
		c.Nodes = append(c.Nodes, node)
		if i == 0 && cfg.RPCPort == 0 {
			// The first node must know its RPCPort or we can't possibly tell
			// the other nodes the correct one to go to.
			//
			// Note: we can't set up a cluster first and clone it for each test,
			// because all ports change so the cluster won't come together.
			// Luckily, it takes only ~2 seconds from zero to a replicated 4
			// node cluster.
			if err := <-chs[0]; err != nil {
				log.Fatalf(ctx, "while starting first node: %s", err)
			}
			ch := make(chan error)
			close(ch)
			chs[0] = ch
		}
	}

	for i := range chs {
		if err := <-chs[i]; err != nil {
			log.Fatalf(ctx, "node %d: %s", i+1, err)
		}
	}

	log.Infof(context.Background(), "started %.3fs", timeutil.Since(c.started).Seconds())
	c.waitForFullReplication()
}

// Close stops the cluster, killing all of the nodes.
func (c *Cluster) Close() {
	for _, n := range c.Nodes {
		n.Kill()
	}
	c.stopper.Stop(context.Background())
	if c.cfg.Ephemeral {
		_ = os.RemoveAll(c.cfg.DataDir)
	}
}

func (c *Cluster) joins() []string {
	type addrAndSeq struct {
		addr string
		seq  int32
	}

	var joins []addrAndSeq
	for _, node := range c.Nodes {
		advertAddr := node.advertiseAddr()
		if advertAddr != "" {
			joins = append(joins, addrAndSeq{
				addr: advertAddr,
				seq:  atomic.LoadInt32(&node.startSeq),
			})
		}
	}
	sort.Slice(joins, func(i, j int) bool {
		return joins[i].seq < joins[j].seq
	})

	if len(joins) == 0 {
		return nil
	}

	// Return the node with the smallest startSeq, i.e. the node that was
	// started first. This is the node that might have no --join flag set, and
	// we must point the other nodes at it, and *only* at it (or the other nodes
	// may connect sufficiently and never bother to talk to this node).
	//
	// See https://github.com/cockroachdb/cockroach/issues/18027 and note that this
	// code is just an unsubstantiated claim. We have observed similar problems with
	// this code. Likely what's described above is only a theoretical problem, but
	// it's good to avoid those problems too.
	//
	// TODO(tschottdorf): revisit after #18027 above closes.
	return []string{joins[0].addr}
}

// IPAddr returns the IP address of the specified node.
func (c *Cluster) IPAddr(nodeIdx int) string {
	return c.Nodes[nodeIdx].IPAddr()
}

// RPCPort returns the RPC port of the specified node. Returns zero if unknown.
func (c *Cluster) RPCPort(nodeIdx int) string {
	return c.Nodes[nodeIdx].RPCPort()
}

// HTTPPort returns the HTTP port of the specified node. Returns zero if unknown.
func (c *Cluster) HTTPPort(nodeIdx int) string {
	return c.Nodes[nodeIdx].HTTPPort()
}

func (c *Cluster) makeNode(ctx context.Context, nodeIdx int, cfg NodeConfig) (*Node, <-chan error) {
	baseCtx := &base.Config{
		User:     security.NodeUser,
		Insecure: true,
	}
	rpcCtx := rpc.NewContext(log.AmbientContext{Tracer: tracing.NewTracer()}, baseCtx,
		hlc.NewClock(hlc.UnixNano, 0), c.stopper)

	node := &Node{
		cfg:    cfg,
		rpcCtx: rpcCtx,
		seq:    c.seq,
	}

	args := []string{
		cfg.Binary,
		"start",
		"--insecure",
		fmt.Sprintf("--host=%s", node.IPAddr()),
		fmt.Sprintf("--port=%d", cfg.RPCPort),
		fmt.Sprintf("--http-port=%d", cfg.HTTPPort),
		fmt.Sprintf("--store=%s", cfg.DataDir),
		fmt.Sprintf("--listening-url-file=%s", node.listeningURLFile()),
		fmt.Sprintf("--cache=256MiB"),
	}

	node.cfg.ExtraArgs = append(args, cfg.ExtraArgs...)

	if err := os.MkdirAll(node.logDir(), 0755); err != nil {
		log.Fatal(context.Background(), err)
	}

	joins := c.joins()
	if nodeIdx > 0 && len(joins) == 0 {
		ch := make(chan error, 1)
		ch <- errors.Errorf("node %d started without join flags", nodeIdx+1)
		return node, ch
	}
	ch := node.StartAsync(ctx, joins...)
	return node, ch
}

// waitForFullReplication waits for the cluster to be fully replicated.
func (c *Cluster) waitForFullReplication() {
	for i := 1; true; i++ {
		done, detail := c.isReplicated()
		if (done && i >= 50) || (i%50) == 0 {
			fmt.Print(detail)
			log.Infof(context.Background(), "waiting for replication")
		}
		if done {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Infof(context.Background(), "replicated %.3fs", timeutil.Since(c.started).Seconds())
}

// Client returns a *client.DB for the node with the given index.
func (c *Cluster) Client(idx int) *client.DB {
	return c.Nodes[idx].Client()
}

func (c *Cluster) isReplicated() (bool, string) {
	db := c.Client(0)
	rows, err := db.Scan(context.Background(), keys.Meta2Prefix, keys.Meta2Prefix.PrefixEnd(), 100000)
	if err != nil {
		if IsUnavailableError(err) {
			return false, ""
		}
		log.Fatalf(context.Background(), "scan failed: %s\n", err)
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)

	done := true
	for _, row := range rows {
		desc := &roachpb.RangeDescriptor{}
		if err := row.ValueProto(desc); err != nil {
			log.Fatalf(context.Background(), "%s: unable to unmarshal range descriptor\n", row.Key)
			continue
		}
		var storeIDs []roachpb.StoreID
		for _, replica := range desc.Replicas {
			storeIDs = append(storeIDs, replica.StoreID)
		}
		fmt.Fprintf(tw, "\t%s\t%s\t[%d]\t%d\n",
			desc.StartKey, desc.EndKey, desc.RangeID, storeIDs)
		// This check is coarse since it doesn't know the real configuration.
		// Assume all is well when there are 3+ replicas, or if there are as
		// many replicas as there are nodes.
		if len(desc.Replicas) < 3 && len(desc.Replicas) != len(c.Nodes) {
			done = false
		}
	}
	_ = tw.Flush()
	return done, buf.String()
}

// UpdateZoneConfig updates the default zone config for the cluster.
func (c *Cluster) UpdateZoneConfig(rangeMinBytes, rangeMaxBytes int64) {
	zone := config.DefaultZoneConfig()
	zone.RangeMinBytes = rangeMinBytes
	zone.RangeMaxBytes = rangeMaxBytes

	buf, err := protoutil.Marshal(&zone)
	if err != nil {
		log.Fatal(context.Background(), err)
	}
	_, err = c.Nodes[0].DB().Exec(`UPSERT INTO system.zones (id, config) VALUES (0, $1)`, buf)
	if err != nil {
		log.Fatal(context.Background(), err)
	}
}

// Split splits the range containing the split key at the specified split key.
func (c *Cluster) Split(nodeIdx int, splitKey roachpb.Key) error {
	return c.Client(nodeIdx).AdminSplit(context.Background(), splitKey, splitKey)
}

// TransferLease transfers the lease for the range containing key to a random
// alive node in the range.
func (c *Cluster) TransferLease(nodeIdx int, r *rand.Rand, key roachpb.Key) (bool, error) {
	desc, err := c.lookupRange(nodeIdx, key)
	if err != nil {
		return false, err
	}
	if len(desc.Replicas) <= 1 {
		return false, nil
	}

	var target roachpb.StoreID
	for {
		target = desc.Replicas[r.Intn(len(desc.Replicas))].StoreID
		if c.Nodes[target-1].Alive() {
			break
		}
	}
	if err := c.Client(nodeIdx).AdminTransferLease(context.Background(), key, target); err != nil {
		return false, errors.Errorf("%s: transfer lease: %s", key, err)
	}
	return true, nil
}

func (c *Cluster) lookupRange(nodeIdx int, key roachpb.Key) (*roachpb.RangeDescriptor, error) {
	req := &roachpb.RangeLookupRequest{
		Span: roachpb.Span{
			Key: keys.RangeMetaKey(keys.MustAddr(key)),
		},
		MaxRanges: 1,
	}
	sender := c.Client(nodeIdx).GetSender()
	resp, pErr := client.SendWrapped(context.Background(), sender, req)
	if pErr != nil {
		return nil, errors.Errorf("%s: lookup range: %s", key, pErr)
	}
	return &resp.(*roachpb.RangeLookupResponse).Ranges[0], nil
}

// RandNode returns the index of a random alive node.
func (c *Cluster) RandNode(f func(int) int) int {
	for {
		i := f(len(c.Nodes))
		if c.Nodes[i].Alive() {
			return i
		}
	}
}

// Node holds the state for a single node in a local cluster and provides
// methods for starting, pausing, resuming and stopping the node.
type Node struct {
	cfg    NodeConfig
	rpcCtx *rpc.Context
	seq    *seqGen

	startSeq int32 // updated atomically on start

	syncutil.Mutex
	cmd                      *exec.Cmd
	rpcPort, httpPort, pgURL string
	db                       *gosql.DB
	client                   *client.DB
	statusClient             serverpb.StatusClient
}

// RPCPort returns the RPC + Postgres port.
func (n *Node) RPCPort() string {
	n.Lock()
	defer n.Unlock()
	return n.rpcPort
}

// RPCAddr returns the RPC + Postgres address, or an empty string if it is not known
// (for instance since the node is down).
func (n *Node) RPCAddr() string {
	port := n.RPCPort()
	if port == "" || port == "0" {
		return ""
	}
	return net.JoinHostPort(n.IPAddr(), port)
}

// HTTPPort returns the ui port (may be empty until known).
func (n *Node) HTTPPort() string {
	n.Lock()
	defer n.Unlock()
	return n.httpPort
}

// PGUrl returns the postgres connection string (may be empty until known).
func (n *Node) PGUrl() string {
	n.Lock()
	defer n.Unlock()
	return n.pgURL
}

// Alive returns true if the node is alive (i.e. not stopped). Note that a
// paused node is considered alive.
func (n *Node) Alive() bool {
	n.Lock()
	defer n.Unlock()
	return n.cmd != nil
}

// Client returns a *client.DB set up to talk to this node.
func (n *Node) Client() *client.DB {
	n.Lock()
	existingClient := n.client
	n.Unlock()

	if existingClient != nil {
		return existingClient
	}

	conn, err := n.rpcCtx.GRPCDial(n.RPCAddr())
	if err != nil {
		log.Fatalf(context.Background(), "failed to initialize KV client: %s", err)
	}
	return client.NewDB(client.NewSender(conn), n.rpcCtx.LocalClock)
}

// StatusClient returns a StatusClient set up to talk to this node.
func (n *Node) StatusClient() serverpb.StatusClient {
	n.Lock()
	existingClient := n.statusClient
	n.Unlock()

	if existingClient != nil {
		return existingClient
	}

	conn, err := n.rpcCtx.GRPCDial(n.RPCAddr())
	if err != nil {
		log.Fatalf(context.Background(), "failed to initialize status client: %s", err)
	}
	return serverpb.NewStatusClient(conn)
}

func (n *Node) logDir() string {
	if n.cfg.LogDir == "" {
		return filepath.Join(n.cfg.DataDir, "logs")
	}
	return n.cfg.LogDir
}

func (n *Node) listeningURLFile() string {
	return filepath.Join(n.cfg.DataDir, listeningURLFile)
}

// Start starts a node.
func (n *Node) Start(ctx context.Context, joins ...string) {
	if err := <-n.StartAsync(ctx, joins...); err != nil {
		log.Fatal(ctx, err)
	}
}

func (n *Node) setNotRunningLocked() {
	_ = os.Remove(n.listeningURLFile())
	_ = os.Remove(n.advertiseAddrFile())
	n.db = nil
	n.client = nil
	n.statusClient = nil
	n.cmd = nil
	n.rpcPort = ""
	n.httpPort = ""
	atomic.StoreInt32(&n.startSeq, 0)
}

func (n *Node) startAsyncInnerLocked(ctx context.Context, joins ...string) error {
	n.setNotRunningLocked()

	args := append([]string(nil), n.cfg.ExtraArgs[1:]...)
	for _, join := range joins {
		args = append(args, "--join", join)
	}
	n.cmd = exec.Command(n.cfg.ExtraArgs[0], args...)
	n.cmd.Env = os.Environ()
	n.cmd.Env = append(n.cmd.Env, n.cfg.ExtraEnv...)

	atomic.StoreInt32(&n.startSeq, n.seq.Next())

	_ = os.MkdirAll(n.logDir(), 0755)

	stdoutPath := filepath.Join(n.logDir(), "stdout")
	stdout, err := os.OpenFile(stdoutPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return errors.Wrapf(err, "unable to open file %s", stdoutPath)
	}
	// This causes the "node startup header" to be printed to stdout, which is
	// helpful and not too noisy.
	n.cmd.Stdout = io.MultiWriter(stdout, os.Stdout)

	stderrPath := filepath.Join(n.logDir(), "stderr")
	stderr, err := os.OpenFile(stderrPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return errors.Wrapf(err, "unable to open file %s", stderrPath)
	}
	n.cmd.Stderr = stderr

	if n.cfg.RPCPort > 0 {
		n.rpcPort = fmt.Sprintf("%d", n.cfg.RPCPort)
	}
	if n.cfg.HTTPPort > 0 {
		n.httpPort = fmt.Sprintf("%d", n.cfg.HTTPPort)
	}

	if err := n.cmd.Start(); err != nil {
		if err := stdout.Close(); err != nil {
			log.Warning(ctx, err)
		}
		if err := stderr.Close(); err != nil {
			log.Warning(ctx, err)
		}
		return errors.Wrapf(err, "running %s %v", n.cmd.Path, n.cmd.Args)
	}

	log.Infof(ctx, "process %d starting: %s", n.cmd.Process.Pid, n.cmd.Args)

	go func(cmd *exec.Cmd) {
		if err := cmd.Wait(); err != nil {
			log.Warning(ctx, err)
		}
		if err := stdout.Close(); err != nil {
			log.Warning(ctx, err)
		}
		if err := stderr.Close(); err != nil {
			log.Warning(ctx, err)
		}

		log.Infof(ctx, "process %d: %s", cmd.Process.Pid, cmd.ProcessState)

		n.Lock()
		n.setNotRunningLocked()
		n.Unlock()
	}(n.cmd)

	return nil
}

// StartAsync starts a node asynchronously. It returns a channel that receives either
// an error, or, once the node has started up and is fully functional, `nil`.
//
// StartAsync is a no-op if the node is already running.
func (n *Node) StartAsync(ctx context.Context, joins ...string) <-chan error {
	ch := make(chan error, 1)

	if err := func() error {
		n.Lock()
		defer n.Unlock()
		if n.cmd != nil {
			return errors.New("server is already running")
		}
		return n.startAsyncInnerLocked(ctx, joins...)
	}(); err != nil {
		ch <- err
		return ch
	}

	isServing := make(chan struct{})
	go func() {
		n.waitUntilLive()
		close(isServing)
		ch <- nil
	}()

	// This blocking loop in the sync path is counter-intuitive but is essential
	// in allowing restarts of whole clusters. Roughly the following happens:
	//
	// 1. The whole cluster gets killed.
	// 2. A node restarts.
	// 3. It will *block* here until it has written down the file which contains
	//    enough information to link other nodes.
	// 4. When restarting other nodes, and `.joins()` is passed in, these nodes
	//    can connect (at least) to the first node.
	// 5. the cluster can become healthy after restart.
	//
	// If we didn't block here, we'd start all nodes up with join addresses that
	// don't make any sense, and the cluster would likely not become connected.
	//
	// An additional difficulty is that older versions (pre 1.1) don't write
	// this file. That's why we let *every* node do this (you could try to make
	// only the first one wait, but if that one is 1.0, bad luck).
	// Short-circuiting the wait in the case that the listening URL file is
	// written (i.e. isServing closes) makes restarts work with 1.0 servers for
	// the most part.
	for {
		if gossipAddr := n.advertiseAddr(); gossipAddr != "" {
			_, port, err := net.SplitHostPort(gossipAddr)
			if err != nil {
				ch = make(chan error, 1)
				ch <- errors.Wrapf(err, "can't parse gossip address %s", gossipAddr)
				return ch
			}
			n.rpcPort = port
			return ch
		}
		select {
		case <-isServing:
			return ch
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func portFromURL(rawURL string) (string, *url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, err
	}

	_, port, err := net.SplitHostPort(u.Host)
	return port, u, err
}

func makeDB(url string, numWorkers int, dbName string) *gosql.DB {
	conn, err := gosql.Open("postgres", url)
	if err != nil {
		log.Fatal(context.Background(), err)
	}
	if numWorkers == 0 {
		numWorkers = 1
	}
	conn.SetMaxOpenConns(numWorkers)
	conn.SetMaxIdleConns(numWorkers)
	return conn
}

func (n *Node) advertiseAddrFile() string {
	return filepath.Join(n.cfg.DataDir, "cockroach.advertise-addr")
}

func (n *Node) advertiseAddr() (s string) {
	c, err := ioutil.ReadFile(n.advertiseAddrFile())
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		// The below is part of the workaround for nodes at v1.0 which don't
		// write the file above, explained in more detail in StartAsync().
		if port := n.RPCPort(); port != "" {
			return net.JoinHostPort(n.IPAddr(), n.RPCPort())
		}
		return ""
	}
	return string(c)
}

func (n *Node) waitUntilLive() {
	ctx := context.Background()
	opts := retry.Options{
		InitialBackoff: time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		Multiplier:     2,
	}
	for r := retry.Start(opts); r.Next(); {
		urlBytes, err := ioutil.ReadFile(n.listeningURLFile())
		if err != nil {
			continue
		}

		var pgURL *url.URL
		_, pgURL, err = portFromURL(string(urlBytes))
		if err != nil {
			log.Info(ctx, err)
			continue
		}

		if n.cfg.RPCPort == 0 {
			n.Lock()
			n.rpcPort = pgURL.Port()
			n.Unlock()
		}

		pgURL.Path = n.cfg.DB
		n.Lock()
		n.pgURL = pgURL.String()
		pid := n.cmd.Process.Pid
		n.Unlock()

		var uiURL *url.URL

		defer func() {
			log.Infof(ctx, "process %d started (db: %s ui: %s)", pid, pgURL, uiURL)
		}()

		// We're basically running, but (at least) the decommissioning test sometimes starts
		// up servers that can already be draining when they get here. For that reason, leave
		// the admin port undefined if we don't manage to get it.
		//
		// This can be improved by making the below code run opportunistically whenever the
		// http port is required but isn't initialized yet.
		n.Lock()
		n.db = makeDB(n.pgURL, n.cfg.NumWorkers, n.cfg.DB)
		n.Unlock()

		{
			var uiStr string
			if err := n.db.QueryRow(
				`SELECT value FROM crdb_internal.node_runtime_info WHERE component='UI' AND field = 'URL'`,
			).Scan(&uiStr); err != nil {
				log.Info(ctx, err)
				break
			}

			n.Lock()
			n.httpPort, uiURL, err = portFromURL(uiStr)
			n.Unlock()
			if err != nil {
				log.Info(ctx, err)
				// TODO(tschottdorf): see above.
			}
		}
		break
	}
}

// Kill stops a node abruptly by sending it SIGKILL.
func (n *Node) Kill() {
	func() {
		n.Lock()
		defer n.Unlock()
		if n.cmd == nil || n.cmd.Process == nil {
			return
		}
		_ = n.cmd.Process.Kill()
	}()
	// Wait for the process to have been cleaned up (or a call to Start() could
	// turn into an unintended no-op).
	for ok := false; !ok; {
		n.Lock()
		ok = n.cmd == nil
		n.Unlock()
	}
}

// IPAddr returns the node's listening address (for ui, inter-node, cli, and
// Postgres alike).
func (n *Node) IPAddr() string {
	return n.cfg.Addr
}

// DB returns a Postgres connection set up to talk to the node.
func (n *Node) DB() *gosql.DB {
	n.Lock()
	defer n.Unlock()
	return n.db
}
