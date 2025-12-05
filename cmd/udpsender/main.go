package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Error while resolving UDP address.")
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("Error while established an udp server.")
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading string")
		}
		fmt.Printf("Entered string: %s", str)

		n, err := conn.Write([]byte(str))
		if err != nil {
			log.Fatal("Error reading string")
		}
		fmt.Printf("Bytes write: %d\n", n)
	}
}
