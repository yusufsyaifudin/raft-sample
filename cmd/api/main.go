package main

import (
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
	"ysf/raftsample/fsm"
	"ysf/raftsample/server"
)

// configRaft configuration for raft node
type configRaft struct {
	NodeId    string `mapstructure:"node_id"`
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
}

// configServer configuration for HTTP server
type configServer struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// config configuration
type config struct {
	Server configServer `mapstructure:"server"`
	Raft   configRaft   `mapstructure:"raft"`
}

const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
)

// main entry point of application start
func main() {
	args := os.Args
	fmt.Println(args)

	conf := config{}

	pflag.String("config", "config.yaml", "select the config file, default to config.yaml")
	pflag.Parse()

	var v = viper.New()
	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatal(err)
		return
	}

	v.SetConfigType("yaml")
	v.SetConfigFile(v.GetString("config"))

	log.Printf("using config file %s to running program\n", v.GetString("config"))

	err := v.ReadInConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = v.Unmarshal(&conf)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Preparing badgerDB
	badgerOpt := badger.DefaultOptions(conf.Raft.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()

	var raftBinAddr = fmt.Sprintf("%s:%d", conf.Raft.Host, conf.Raft.Port)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.Raft.NodeId)
	raftConf.SnapshotThreshold = 1024

	fsmStore := fsm.NewBadger(badgerDB)

	store, err := raftboltdb.NewBoltStore(filepath.Join(conf.Raft.VolumeDir, "raft.dataRepo"))
	if err != nil {
		log.Fatal(err)
		return
	}

	// Wrap the store in a LogCache to improve performance.
	cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		log.Fatal(err)
		return
	}

	snapshotStore, err := raft.NewFileSnapshotStore(conf.Raft.VolumeDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		log.Fatal(err)
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBinAddr)
	if err != nil {
		log.Fatal(err)
		return
	}

	transport, err := raft.NewTCPTransport(raftBinAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		log.Fatal(err)
		return
	}

	raftServer, err := raft.NewRaft(raftConf, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		log.Fatal(err)
		return
	}

	// always start single server as a leader
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(conf.Raft.NodeId),
				Address: transport.LocalAddr(),
			},
		},
	}

	raftServer.BootstrapCluster(configuration)

	srv := server.New(fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port), badgerDB, raftServer)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	return
}
