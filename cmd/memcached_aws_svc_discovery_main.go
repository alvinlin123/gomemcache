package main

import (
	"fmt"

	"../memcache"
)


func main() {
	ecsd := memcache.NewElastiCacheServiceDiscovery("localhost:11211")
	nodes, err := ecsd.GetNodes()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("version %v", nodes)
}