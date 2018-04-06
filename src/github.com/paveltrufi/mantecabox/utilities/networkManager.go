package utilities

import (
	"net"
	"fmt"
	"strings"
)

func GetIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		fmt.Print(err)
	}

	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}
