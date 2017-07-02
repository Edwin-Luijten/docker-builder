package main

import (
	"fmt"
	"strconv"
)

const listener_port int = 9001
const interface_port int = 8080

func main() {
	fmt.Println("automated docker builds")
	fmt.Println("webhook listener listens on port: " + strconv.Itoa(listener_port));
	fmt.Println("interface listens on port: " + strconv.Itoa(interface_port));
}