package cluster

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

const (
	retainSnapshotCount = 2
)

//Rnode implement raft consensus
type Rnode struct {
	raftDir  string
	raftBind string
	nid      string
	raft     *raft.Raft
	logger   log.Logger
	meta     map[string]string //for raft-save
	metaMu   sync.Mutex
}

// NewRnode returns a newObject
func NewRnode(dir string, addr string, id string, lg log.Logger) *Rnode {
	n := &Rnode{
		raftDir:  dir,
		raftBind: addr,
		nid:      id,
		logger:   lg,
	}
	n.meta = make(map[string]string)
	return n
}

func (r *Rnode) Start(single bool) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(r.nid)
	addr, err := net.ResolveTCPAddr("tcp", r.raftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(r.raftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}
	snapshots, err := raft.NewFileSnapshotStore(r.raftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(r.raftDir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}
	ra, err := raft.NewRaft(config, r, boltDB, boltDB, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	r.raft = ra
	if single {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}
	return nil
}
