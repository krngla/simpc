package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func handleConnection(conn net.Conn) {

	defer conn.Close()
	//writer.WriteString("END\n")
	done := make(chan bool)
	readerr := make(chan error)
	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("dialing:", err.Error())
				break
			}
			fmt.Printf("%s\n", line)
			time.Sleep(1 * time.Second)
		}
		done <- true
		readerr <- nil
	}()

	writer := bufio.NewWriter(conn)
	ioreader := bufio.NewReader(os.Stdin)
	for {
		line, err := ioreader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("dialing:", err.Error())
			break
		}
		writer.WriteString(line)
		writer.Flush()
	}
	<-done
	err := <-readerr
	if err != nil {
		log.Fatal("dialing:", err.Error())
	}
	fmt.Printf("close connection: %v\n", conn)
}

func main() {
	s := simpc.NewServer(handleConnection, simpc.defaultPath)
	s.HandleConnection = handleConnection
	s.ListenAndServe()

}
