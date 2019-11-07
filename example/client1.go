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
	gocache.Pool.Put("key1",1,60 * time.Second)

	time.Sleep(60 * time.Second)

}