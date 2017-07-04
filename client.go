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
	Identifier string
	Type       string
	Job        string
	Message    string
	Error      string
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

	// Send back a response
	conn.Write([]byte("Update received"))

	// Decode json to object
	err := request.Decode(&message)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Identifier", message.Identifier)
	log.Println("Type: " + message.Type)
	log.Println("Job: " + message.Job)
	log.Println("Message: " + message.Message)
	log.Println("Error: " + message.Error)

	conn.Close()
}

func getIp() (string) {

	address, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	var ip string

	for _, a := range address {

		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
			}
		}
	}

	return ip
}
