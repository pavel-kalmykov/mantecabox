package utilities

import (
	"log"
	"net"
	"strings"
)

func GetIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		log.Println(err)
	}

	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}
