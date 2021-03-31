package kvraft

import (
	"6.824/src/labrpc"
	"sync"
	"sync/atomic"
	"time"
)
import "crypto/rand"
import "math/big"

type Clerk struct {
	servers []*labrpc.ClientEnd
	// You will have to modify this struct.
	mu         sync.Mutex
	leaderId   int
	clientId   int64
	sequenceId int64
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	// You'll have to add code here.
	ck.clientId = nrand()
	return ck
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Get(key string) string {
	// You will have to modify this function.
	args := GetArgs{Key: key, ClientId: ck.clientId, SequenceId: atomic.AddInt64(&ck.sequenceId, 1)}
	DPrintf("Client[%d] Get starts, Key=%s ", ck.clientId, key)
	leaderId := ck.currentLeader()
	for {
		reply := GetReply{}
		if ck.servers[leaderId].Call("KVServer.Get", &args, &reply) {
			if reply.Err == OK {
				return reply.Value
			} else if reply.Err == ErrNoKey {
				return ""
			}
		}
		leaderId = ck.changeLeader()
		time.Sleep(1 * time.Millisecond)
	}
}

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) PutAppend(key string, value string, op string) {
	args := PutAppendArgs{
		// You will have to modify this function.
		Key:        key,
		Value:      value,
		Op:         op,
		SequenceId: atomic.AddInt64(&ck.sequenceId, 1),
		ClientId:   ck.clientId,
	}
	leaderId := ck.currentLeader()
	for {
		reply := PutAppendReply{}
		if ck.servers[leaderId].Call("KVServer.PutAppend", &args, &reply) {
			if reply.Err == OK {
				break
			}
		}
		leaderId = ck.changeLeader()
		time.Sleep(1 * time.Millisecond)
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}

func (ck *Clerk) currentLeader() (leaderId int) {
	ck.mu.Lock()
	defer ck.mu.Unlock()
	leaderId = ck.leaderId
	return
}

func (ck *Clerk) changeLeader() (leaderId int) {
	ck.mu.Lock()
	defer ck.mu.Unlock()
	ck.leaderId = (ck.leaderId + 1) % len(ck.servers)
	return ck.leaderId
}
