package main

import (
	"fmt"
	"strconv"
	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Title string
	Ports ports
}

type ports struct {
	updates	int
	ui		int
}

func main() {

	var config Configuration

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(config.Title)
	fmt.Println("webhook listener listens on port: " + strconv.Itoa(config.Ports.updates));
	fmt.Println("interface listens on port: " + strconv.Itoa(config.Ports.ui));
}