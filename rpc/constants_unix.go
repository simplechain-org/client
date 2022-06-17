
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package rpc

import "C"

var (
	max_path_size = C.max_socket_path_size()
)
