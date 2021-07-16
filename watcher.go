// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package watcher

import (
	"bytes"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
	goctx "golang.org/x/net/context"
)

type Watcher struct {
	sync.Mutex
	client       kv.Storage
	pollInterval time.Duration
	callback     func(string) //should be protected by lock
	keyName      string
	formerValue  []byte //should be protected by lock
	done         bool   //should be protected by lock
}

func NewWatcher(keyName string, pollInterval time.Duration, pdAddrs string) (*Watcher, error) {
	cli, err := tikv.Driver{}.Open("tikv://" + pdAddrs)
	if err != nil {
		return nil, err
	}

	res := Watcher{
		client:       cli,
		callback:     nil,
		keyName:      keyName,
		pollInterval: pollInterval,
		formerValue:  make([]byte, 0),
		done:         false,
	}
	go res.poll()
	return &res, nil
}

// SetUpdateCallback sets the callback function that the watcher will call
// when the policy in DB has been changed by other instances.
// A classic callback is Enforcer.LoadPolicy().
func (w *Watcher) SetUpdateCallback(callback func(string)) error {
	w.Lock()
	defer w.Unlock()
	w.callback = callback
	return nil
}

// Update calls the update callback of other instances to synchronize their policy.
// It is usually called after changing the policy in DB, like Enforcer.SavePolicy(),
// Enforcer.AddPolicy(), Enforcer.RemovePolicy(), etc.
func (w *Watcher) Update() error {
	//begin transaction
	tx, err := w.client.Begin()
	if err != nil {
		return err
	}
	rev := 0
	val, err := tx.Get(goctx.Background(), []byte(w.keyName))
	if err != nil {
		if err != kv.ErrNotExist {
			return err
		}
	} else {
		rev, err = strconv.Atoi(string(val))
		if err != nil {
			return err
		}
		log.Println("Get revision: ", rev)
		rev += 1
	}
	log.Printf("set revision %d\n", rev)
	if err = tx.Set([]byte(w.keyName), []byte(strconv.Itoa(rev))); err != nil {
		return err
	}
	w.Lock()
	if err = tx.Commit(goctx.Background()); err != nil {
		return err
	}

	w.formerValue = []byte(strconv.Itoa(rev))
	w.Unlock()
	return err

}

func (w *Watcher) Close() {
	w.Lock()
	defer w.Unlock()
	w.done = true
}

// startWatch is a goroutine that watches the policy change by polling.
func (w *Watcher) poll() {
	for {
		time.Sleep(w.pollInterval)
		w.Lock()
		if w.done {
			w.Unlock()
			break
		}
		w.Unlock()
		tx, err := w.client.Begin()
		if err != nil {
			log.Printf("Watcher polling goroutine exited due to error: %s\n", err)
			return
		}
		val, err := tx.Get(goctx.Background(), []byte(w.keyName))
		tx.Commit(goctx.Background())
		if err != nil && err != kv.ErrNotExist {
			log.Printf("Watcher polling goroutine exited due to error: %s\n", err)
			return
		} else if err == kv.ErrNotExist {
			val = nil
		}

		w.Lock()
		//fmt.Printf("Get %s,former %s\n",string(val),string(w.formerValue))
		if !bytes.Equal(w.formerValue, val) {
			var callback = w.callback
			w.formerValue = val
			w.Unlock()
			if callback != nil {
				callback(string(val))
			}

		} else {
			w.Unlock()
		}
	}

}
