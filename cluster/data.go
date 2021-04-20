package cluster

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

//下面函数的实现都和内存中存储结构息息相关

// Apply applies a Raft log entry to the key-value store.
func (r *Rnode) Apply(l *raft.Log) interface{} {

}

// Snapshot returns a snapshot of the database. The caller must ensure that
// no transaction is taking place during this call. Hashicorp Raft guarantees
// that this function will not be called concurrently with Apply, as it states
// Apply and Snapshot are always called from the same thread. This means there
// is no need to synchronize this function with Execute(). However queries that
// involve a transaction must be blocked.
//
func (f *Rnode) Snapshot() (raft.FSMSnapshot, error) {
	fsm := &fsmSnapshot{}	
	fsm.meta=
}

// Restore stores the key-value store to a previous state.
func (f *Rnode) Restore(rc io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	f.meta = o
	return nil
}

//inplement raft FSMSnapshot interface
type fsmSnapshot struct {
	meta []byte
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		// Write data to sink.
		if _, err := sink.Write(f.meta); err != nil {
			return err
		}
		// Close the sink.
		return sink.Close()
	}()
	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

// Release is a no-op.
func (f *fsmSnapshot) Release() {}
