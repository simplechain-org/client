
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package rpc

/*
#include <sys/un.h>

int client_max_socket_path_size() {
struct sockaddr_un s;
return sizeof(s.sun_path);
}

*/
import "C"

var (
	max_path_size = C.client_max_socket_path_size()
)
