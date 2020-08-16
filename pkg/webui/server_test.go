// +build integration

package webui

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestServer_Serve(t *testing.T) {
	srv := NewServer("127.0.0.1:0", "", "", nil, logrus.NewEntry(logrus.New()))
	ctx, cancelFn := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := srv.Serve(ctx)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	cancelFn()
	wg.Wait()
}
