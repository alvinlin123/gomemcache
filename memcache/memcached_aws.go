package memcache

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type ElastiCacheServiceDiscovery struct {
	client *Client
	configEndpoint string
}

type CacheNodes struct {
	version uint64
	nodes []CacheNode
}

type CacheNode struct {
	dns string
	ip string
	port int
}

func NewElastiCacheServiceDiscovery(configEndpoint string) *ElastiCacheServiceDiscovery {
	client := New(configEndpoint)
	client.MaxIdleConns = 0
	client.Timeout = 10 * time.Second
	return &ElastiCacheServiceDiscovery{
		configEndpoint: configEndpoint,
		client: client,
	}
}

func (ecsd *ElastiCacheServiceDiscovery) GetNodes() (*CacheNodes, error) {
	configSrvAddr, err := ecsd.client.selector.PickServer("any")
	if err != nil {
		return nil, err
	}

	result := &CacheNodes{}

	err = ecsd.client.withAddrRw(configSrvAddr, func(rw *bufio.ReadWriter) error {
		if _, err := fmt.Fprintf(rw, "config get cluster\r\n"); err != nil {
			return err
		}
		if err := rw.Flush(); err != nil {
			return err
		}
		reader := rw.Reader
		metaLine, err := reader.ReadSlice('\n')
		if err != nil {
			return err
		}
		metaLineStr := strings.TrimSpace(string(metaLine))

		//Should look like "CONFIG cluster 0 147"
		responseSize, err := strconv.Atoi(strings.Split(metaLineStr, " ")[3])
		if err != nil {
			return err
		}

		buf := make([]byte, responseSize)

		_, err = io.ReadFull(reader, buf)
		if err != nil {
			return err
		}

		lines := strings.Split(string(buf), "\n")

		result.version, err = strconv.ParseUint(strings.TrimSpace(lines[0]), 10, 64)
		hostLines := strings.Split(lines[1], " ")

		for _, h := range(hostLines) {
			info := strings.Split(h, "|")
			port, err := strconv.ParseInt(info[2], 10, 32)
			if err != nil {
				return err
			}
			n := CacheNode {
				dns: info[0],
				ip: info[1],
				port: int(port),
			}

			result.nodes = append(result.nodes, n)
		}

		return nil
	})

	return result, err
}