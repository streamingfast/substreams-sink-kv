package wasmquery

import (
	"testing"
	"time"

	"github.com/test-go/testify/require"
)

func Test_ConnectWebLaunch(t *testing.T) {

	protoFilePath := "./testdata/test.proto"
	service := "sf.test.v1.TestService"

	filedesc := protoFileToDescriptor(t, protoFilePath)

	config, err := NewServiceConfig(filedesc, service)
	require.NoError(t, err)

	c, err := newServer(nil, config, zlog)
	require.NoError(t, err)
	go c.Launch("localhost:8000")

	end := make(chan bool, 1)
	go func() {
		time.Sleep(30 * time.Minute)
		end <- true
	}()

	<-end
}
