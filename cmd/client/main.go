// myipc
package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func handleConnection(conn net.Conn, str string) {
	line, err := reader.ReadString('\n')
	line = strings.TrimSuffix(line, "\n")
	//writer.WriteString(line)
	if err != nil {
		log.Fatal("dialing:", err.Error())
		break
	}
	if strings.Contains(line, "END") {
		break
	}
	fmt.Printf("%s\n", line)
}
