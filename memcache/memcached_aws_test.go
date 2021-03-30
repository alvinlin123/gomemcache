package memcache

import (
	"bufio"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const serverAddr = "127.0.0.1:7357"

type cmdAndResp struct {
	command string
	response string
}

func startServer(data []cmdAndResp) (net.Listener, error) {
	l, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		for _, d := range data {
			rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
			givenCommand := make([]byte, len(d.command))
			_, err := io.ReadFull(rw, givenCommand)
			if err != nil {
				conn.Close()
				return
			}

			if string(givenCommand) != d.command {
				conn.Close()
				return
			}

			rw.Write([]byte(d.response))
			rw.Flush()
		}
	}()

	time.Sleep(1 * time.Second) //give some time for go routine to start accepting connection
	return l, nil
}

func  TestCanExtractNodeInfo(t *testing.T) {
	data := []cmdAndResp {
		{
			command: "config get cluster\r\n",
			response: "CONFIG cluster 0 136\r\n" +
				"12\n" +
				"myCluster.pc4ldq.0001.use1.cache.amazonaws.com|10.82.235.120|11211 myCluster.pc4ldq.0002.use1.cache.amazonaws.com|10.80.249.27|11211\n\r\n" +
				"END\r\n",

		},
	}
	l, err := startServer(data)
	require.NoError(t, err)
	defer l.Close()

	sut := NewElastiCacheServiceDiscovery(serverAddr)
	nodes, err := sut.GetNodes()
	require.NoError(t, err)
	require.Len(t, nodes.nodes, 2, )

	require.Equal(t, "myCluster.pc4ldq.0001.use1.cache.amazonaws.com", nodes.nodes[0].dns)
	require.Equal(t, "10.82.235.120", nodes.nodes[0].ip)
	require.Equal(t, 11211, nodes.nodes[0].port)

	require.Equal(t, "myCluster.pc4ldq.0002.use1.cache.amazonaws.com", nodes.nodes[1].dns)
	require.Equal(t, "10.80.249.27", nodes.nodes[1].ip)
	require.Equal(t, 11211, nodes.nodes[1].port)
}