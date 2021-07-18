module github.com/casbin/tikv-watcher

go 1.16

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace google.golang.org/grpc v1.29.1 => google.golang.org/grpc v1.26.0

require (
	github.com/casbin/casbin/v2 v2.33.0
	github.com/pingcap/tidb v1.1.0-beta.0.20210419034717-00632fb3c710
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc
)
