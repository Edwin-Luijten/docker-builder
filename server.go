package main

import (
	"net"
	"os"
	"flag"
	"log"
	"strconv"
	"encoding/json"
)

type messageTemplate struct {
	Job     string
	Message string
	Error string
}

func main() {
	updatesPort := flag.Int("port", 9001, "The port to send updates to")

	address := getIp() + ":" + strconv.Itoa(*updatesPort)

	listener, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatal("Error listening on:" + address)
	}

	defer listener.Close()

	log.Println("Listening on: " + address)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	request := json.NewDecoder(conn)

	var message messageTemplate

	// Send back an response
	conn.Write([]byte("Update received"))

	// Decode json to object
	err :=request.Decode(&message)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Job: " + message.Job)
	log.Println("Message: " + message.Message)
	log.Println("Error: " + message.Error)

	conn.Close()
}

func getIp() (string) {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	var address string

	for _, a := range addrs {

		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				address = ipnet.IP.String()
			}
		}
	}

	return address
}
