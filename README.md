# TiKV Watcher

[![Go](https://github.com/casbin/tikv-watcher/actions/workflows/go.yml/badge.svg)](https://github.com/casbin/tikv-watcher/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/tikv-watcher)](https://goreportcard.com/report/github.com/casbin/tikv-watcher)
[![Go Reference](https://pkg.go.dev/badge/github.com/casbin/tikv-watcher.svg)](https://pkg.go.dev/github.com/casbin/tikv-watcher)

Tikv-Watcher is the tikv watcher for casbin. With this library, Casbin can synchronize the policy with the database in multiple enforcer instances.

*Note: Considering that tikv doesn't have watch mechanism like etcd or channel like redis, this ugly implementation uses polling to achieve monitoring a certain key, which may cause some performance trouble*

### Installation: 
```shell
go get github.com/casbin/tikv-watcher
```

### Single Example:
start the tikv service before run this example
```golang
package main

import (
	"fmt"
	"time"

	watcher "github.com/casbin/tikv-watcher"
	casbin "github.com/casbin/casbin/v2"
)

func main() {
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	w, err := watcher.NewWatcher(
		"testkey",
		100*time.Millisecond,
		"127.0.0.1:2379",
	)
	if err != nil {
		panic(err)
		return
	}
	e.SetWatcher(w)
	w.SetUpdateCallback(
		func(s string) {
			fmt.Println("===================get" + s)
		},
	)
	e.SavePolicy()
	time.Sleep(10 * time.Second)

}


```
### License:
This project is under Apache 2.0 License. See the LICENSE file for the full license text.
