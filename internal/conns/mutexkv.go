// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"sync"
)

// GlobalMutexKV is a global MutexKV for use within this plugin.
var GlobalMutexKV = newMutexKV()

// mutexKV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
type mutexKV struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

// Locks the mutex for the given key. Caller is responsible for calling Unlock
// for the same key
func (m *mutexKV) Lock(key string) {
	m.get(key).Lock()
}

// Unlock the mutex for the given key. Caller must have called Lock for the same key first
func (m *mutexKV) Unlock(key string) {
	m.get(key).Unlock()
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *mutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}

// Returns a properly initialized MutexKV
func newMutexKV() *mutexKV {
	return &mutexKV{
		store: make(map[string]*sync.Mutex),
	}
}
