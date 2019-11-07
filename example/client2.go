package main

import (
	"github.com/imiskolee/gocache"
	"time"
)

func main() {
	gocache.Init(&gocache.Opt{
		BroadcastAddr: "192.168.1.255",
		BroadcastPort: 8900,
	})
	gocache.Pool.Put("key2",1,60 * time.Second)
	_ = gocache.Pool.Delete("key1")
	time.Sleep(60 * time.Second)
}
