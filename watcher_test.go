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
	"fmt"
	"sync"
	"testing"
	"time"

	casbin "github.com/casbin/casbin/v2"
	"github.com/pingcap/tidb/store/tikv"
	goctx "golang.org/x/net/context"
)

func TestWatcher(t *testing.T) {
	cli, _ := tikv.Driver{}.Open(fmt.Sprintf("tikv://%s", "127.0.0.1:2379"))
	tx, _ := cli.Begin()
	tx.Delete([]byte("testkey"))
	tx.Commit(goctx.Background())

	w1, _ := NewWatcher(
		"testkey",
		100*time.Millisecond,
		"127.0.0.1:2379",
	)
	w2, _ := NewWatcher(
		"testkey",
		100*time.Millisecond,
		"127.0.0.1:2379",
	)
	var mu sync.Mutex
	var called1 = false
	var called2 = false

	updateCallback1 := func(val string) {
		if val == "1" {
			mu.Lock()
			called1 = true
			mu.Unlock()
		}

	}
	updateCallback2 := func(val string) {
		if val == "0" {
			mu.Lock()
			called2 = true
			mu.Unlock()
		}
	}
	w1.SetUpdateCallback(updateCallback1)
	w2.SetUpdateCallback(updateCallback2)
	w1.Update()
	time.Sleep(2 * time.Second)
	mu.Lock()
	if !called2 {
		t.Errorf("updateCallback2 is not correctly called")
	}
	mu.Unlock()
	w2.Update()
	time.Sleep(2 * time.Second)
	mu.Lock()
	if !called1 {
		t.Errorf("updateCallback1 is not correctly called")
	}
	mu.Unlock()
}

func TestWithCasbin(t *testing.T) {
	cli, _ := tikv.Driver{}.Open(fmt.Sprintf("tikv://%s", "127.0.0.1:2379"))
	tx, _ := cli.Begin()
	tx.Delete([]byte("testkey"))
	tx.Commit(goctx.Background())

	w1, _ := NewWatcher(
		"testkey",
		100*time.Millisecond,
		"127.0.0.1:2379",
	)
	w2, _ := NewWatcher(
		"testkey",
		100*time.Millisecond,
		"127.0.0.1:2379",
	)
	var mu sync.Mutex
	var called1 = false
	var called2 = false

	updateCallback1 := func(val string) {
		if val == "1" {
			mu.Lock()
			called1 = true
			mu.Unlock()
		}

	}
	updateCallback2 := func(val string) {
		if val == "0" {
			mu.Lock()
			called2 = true
			mu.Unlock()
		}
	}
	e1, _ := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	e2, _ := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	e1.SetWatcher(w1)
	e2.SetWatcher(w2)
	w1.SetUpdateCallback(updateCallback1)
	w2.SetUpdateCallback(updateCallback2)

	e1.SavePolicy()
	time.Sleep(2 * time.Second)
	mu.Lock()
	if !called2 {
		t.Errorf("updateCallback2 is not correctly called")
	}
	mu.Unlock()
	e2.SavePolicy()
	time.Sleep(2 * time.Second)
	mu.Lock()
	if !called1 {
		t.Errorf("updateCallback1 is not correctly called")
	}
	mu.Unlock()

}
