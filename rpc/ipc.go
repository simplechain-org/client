package rpc

import (
	"context"
	"fmt"
	"net"
)

func IsTemporaryError(err error) bool {
	tempErr, ok := err.(interface {
		Temporary() bool
	})
	return ok && tempErr.Temporary() || isPacketTooBig(err)
}
func isPacketTooBig(err error) bool {
	return false
}

// ServeListener accepts connections on l, serving JSON-RPC on them.
func (s *Server) ServeListener(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if IsTemporaryError(err) {
			fmt.Println("RPC accept error", "err", err)
			continue
		} else if err != nil {
			return err
		}
		fmt.Println("Accepted RPC connection", "conn", conn.RemoteAddr())
		go s.ServeCodec(NewCodec(conn), 0)
	}
}

// DialIPC create a new IPC client that connects to the given endpoint. On Unix it assumes
// the endpoint is the full path to a unix socket, and Windows the endpoint is an
// identifier for a named pipe.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
func DialIPC(ctx context.Context, endpoint string) (*Client, error) {
	return newClient(ctx, func(ctx context.Context) (ServerCodec, error) {
		conn, err := newIPCConnection(ctx, endpoint)
		if err != nil {
			return nil, err
		}
		return NewCodec(conn), err
	})
}
