package rpc

import (
	"github.com/simplechain-org/client/safe"
	"net"
)
// StartHTTPEndpoint starts the HTTP RPC endpoint, configured with cors/vhosts/modules
func StartHTTPEndpoint(endpoint string, apis []API, modules []string, cors []string, vhosts []string, timeouts HTTPTimeouts) (net.Listener, *Server, error) {
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := NewServer()
	for _, api := range apis {
		if whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}
		}
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return nil, nil, err
	}
	go NewHTTPServer(cors, vhosts, timeouts, handler).Serve(listener)
	return listener, handler, err
}
// StartHTTPSEndpoint starts the HTTP RPC endpoint, configured with cors/vhosts/modules
func StartHTTPSEndpoint(endpoint string, certFile string,keyFile string,Type string,certFiles []string,apis []API, modules []string, cors []string, vhosts []string, timeouts HTTPSTimeouts) (net.Listener, *Server, error) {
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := NewServer()
	for _, api := range apis {
		if whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}

		}
	}
	var (
		listener net.Listener
		err      error
	)
	if listener, err = safe.NewTLSServerListener(endpoint,certFile,keyFile,Type,certFiles); err != nil {
		return nil, nil, err
	}
	go NewHTTPSServer(cors, vhosts, timeouts, handler).Serve(listener)
	return listener, handler, err
}
// StartWSEndpoint starts a websocket endpoint
func StartWSEndpoint(endpoint string, apis []API, modules []string, wsOrigins []string, exposeAll bool) (net.Listener, *Server, error) {

	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := NewServer()
	for _, api := range apis {
		if exposeAll || whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}

		}
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return nil, nil, err
	}
	go NewWSServer(wsOrigins, handler).Serve(listener)
	return listener, handler, err

}
// StartWSSEndpoint starts a websocket endpoint
func StartWSSEndpoint(endpoint string, certFile string,keyFile string,Type string,certFiles []string,apis []API, modules []string, wsOrigins []string, exposeAll bool) (net.Listener, *Server, error) {
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := NewServer()
	for _, api := range apis {
		if exposeAll || whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}
		}
	}
	var (
		listener net.Listener
		err      error
	)
	if listener, err = safe.NewTLSServerListener(endpoint,certFile,keyFile,Type,certFiles); err != nil {
		return nil, nil, err
	}
	go NewWSSServer(wsOrigins, handler).Serve(listener)
	return listener, handler, err

}
// StartIPCEndpoint starts an IPC endpoint.
func StartIPCEndpoint(ipcEndpoint string, apis []API) (net.Listener, *Server, error) {
	// Register all the APIs exposed by the services.
	handler := NewServer()
	for _, api := range apis {
		if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
			return nil, nil, err
		}
	}
	// All APIs registered, start the IPC listener.
	listener, err := ipcListen(ipcEndpoint)
	if err != nil {
		return nil, nil, err
	}
	go handler.ServeListener(listener)
	return listener, handler, nil
}
