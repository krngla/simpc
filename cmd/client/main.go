// myipc
package main

import (
	"fmt"
	"time"
)

/*
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
*/

func main() {

	done := make(chan struct{})
	ret := make(chan bool)
	fn := func(done chan struct{}, ret chan bool) {
		for {
			select {
			case <-done:
				ret <- true
				return
			default:
				fmt.Printf("default\n")
			}
		}
	}
	go fn(done, ret)

	time.Sleep(1 * time.Second)
	done <- struct{}{}
	<-ret
	fmt.Printf("\"done\": %v\n", "done")

}
