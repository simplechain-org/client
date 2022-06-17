
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package rpc

/*
#include <sys/un.h>
#ifndef C_MAX_SOCKET_PATH_SIZE_H
#define C_MAX_SOCKET_PATH_SIZE_H
int max_socket_path_size() {
struct sockaddr_un s;
return sizeof(s.sun_path);
}
#endif //C_MAX_SOCKET_PATH_SIZE_H
*/
import "C"

var (
	max_path_size = C.max_socket_path_size()
)
