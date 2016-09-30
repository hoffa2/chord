package netutils

import (
	"sync"

	"github.com/hoffa2/chord/comm"
	"github.com/hoffa2/chord/util"
)

// Wraps the RPC communication
type Remote struct {
	conns map[string]*NodeRPC
	sync.Mutex
}

func NewRemote() *Remote {
	return &Remote{
		conns: make(map[string]*NodeRPC),
	}
}

func (r *Remote) get(host string) (*NodeRPC, error) {
	r.Lock()
	defer r.Unlock()
	val, ok := r.conns[host]
	if !ok {
		conn, err := connectRPC(host)
		if err != nil {
			return nil, err
		}
		r.conns[host] = conn
		return conn, nil
	}
	return val, nil
}

func (r *Remote) GetSuccessor(rn comm.Rnode) (*comm.Rnode, error) {
	c, err := r.get(rn.IP)
	if err != nil {
		return nil, err
	}

	//args := &comm.NodeID{ID: []byte(id)}
	var reply comm.NodeID
	err = c.Call("NodeComm.GetSuccessor", &comm.Empty{}, &reply)
	if err != nil {
		return nil, err
	}
	return &comm.Rnode{ID: util.StringToID(reply.ID), IP: reply.IP}, nil
}

func (r *Remote) GetPredecessor(rn comm.Rnode) (*comm.Rnode, error) {
	c, err := r.get(rn.IP)
	if err != nil {
		return nil, err
	}
	//args := &comm.NodeID{ID: []byte(id)}
	var reply comm.NodeID
	err = c.Call("NodeComm.GetPredecessor", &comm.Empty{}, &reply)
	if err != nil {
		return nil, err
	}
	return &comm.Rnode{ID: util.StringToID(reply.ID), IP: reply.IP}, nil
}

func (r *Remote) FindPredecessor(rn comm.Rnode, id util.Identifier) (*comm.Rnode, error) {
	c, err := r.get(rn.IP)
	if err != nil {
		return nil, err
	}

	args := &comm.NodeID{ID: string(id)}
	var reply comm.NodeID
	err = c.Call("NodeComm.FindPredecessor", args, &reply)
	if err != nil {
		return nil, err
	}
	return &comm.Rnode{ID: util.StringToID(reply.ID), IP: reply.IP}, nil
}

func (r *Remote) FindSuccessor(rn comm.Rnode, id util.Identifier) (*comm.Rnode, error) {
	c, err := r.get(rn.IP)
	if err != nil {
		return nil, err
	}

	args := &comm.NodeID{ID: string(id)}
	var reply comm.NodeID
	err = c.Call("NodeComm.FindSuccessor", args, &reply)
	if err != nil {
		return nil, err
	}
	return &comm.Rnode{ID: util.StringToID(reply.ID), IP: reply.IP}, nil
}

// PutRemote Stores a value in its respective node
func (r *Remote) PutRemote(rn comm.Rnode, key, value string) error {
	c, err := r.get(rn.IP)
	if err != nil {
		return err
	}

	args := &comm.KeyValue{Key: key, Value: value}
	err = c.Call("NodeComm.PutRemote", args, nil)
	if err != nil {
		return err
	}
	return nil
}

// PutRemote Stores a value in its respective node
func (r *Remote) GetRemote(rn comm.Rnode, key string) (string, error) {
	c, err := r.get(rn.IP)
	if err != nil {
		return "", err
	}
	args := &comm.KeyValue{Key: key}
	reply := comm.KeyValue{}
	err = c.Call("NodeComm.GetRemote", args, &reply)
	if err != nil {
		return "", err
	}

	return reply.Value, nil
}

func (r *Remote) UpdatePredecessor(rn comm.Rnode, id util.Identifier, ip string) error {
	c, err := r.get(rn.IP)
	if err != nil {
		return err
	}
	args := &comm.NodeID{ID: string(id), IP: ip}
	err = c.Call("NodeComm.UpdatePredecessor", args, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *Remote) UpdateSuccessor(rn comm.Rnode, id util.Identifier, ip string) error {
	c, err := r.get(rn.IP)
	if err != nil {
		return err
	}

	args := &comm.NodeID{ID: string(id), IP: ip}
	err = c.Call("NodeComm.UpdateSuccessor", args, nil)
	if err != nil {
		return err

	}
	return nil
}

func (r *Remote) ClosestPreFinger(rn comm.Rnode, id util.Identifier) (*comm.Rnode, error) {
	c, err := r.get(rn.IP)
	if err != nil {
		return nil, err
	}

	args := id.ToString()
	var reply comm.NodeID
	err = c.Call("NodeComm.ClosestPreFinger", &args, &reply)
	if err != nil {
		return nil, err
	}

	return &comm.Rnode{ID: util.StringToID(reply.ID), IP: reply.IP}, nil
}

func (r *Remote) UpdateFingerTable(rn comm.Rnode, s util.Identifier, ip string, idx int) error {
	c, err := r.get(rn.IP)
	if err != nil {
		return err
	}

	args := &comm.FingerEntry{
		S:   comm.NodeID{ID: s.ToString(), IP: ip},
		IDX: idx,
	}

	err = c.Call("NodeComm.UpdateFingerTable", args, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *Remote) GetKeysInInterval(rn comm.Rnode, from, to util.Identifier) (*map[string]string, error) {
	c, err := r.get(rn.IP)
	if err != nil {
		return nil, err
	}

	args := &comm.Interval{
		From: from.ToString(),
		To:   to.ToString(),
	}

	reply := make(map[string]string)
	err = c.Call("NodeComm.GetKeysInInterval", args, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

func (r *Remote) Notify(rn comm.Rnode, node *comm.Rnode) error {
	c, err := r.get(rn.IP)
	if err != nil {
		return nil, err
	}

	err = c.Call("NodeComm.Notify", node, &comm.Empty{})
	if err != nil {
		return nil, err
	}
	return &reply, nil
}
