package main

import (
	"fmt"
	"strconv"
	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Title string
	Ports ports `toml:"ports"`
}

type ports struct {
	webHook	int
	ui	int
}

func main() {

	var config Configuration

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(config.Title)
	fmt.Println("webhook listener listens on port: " + strconv.Itoa(config.Ports.webHook));
	fmt.Println("interface listens on port: " + strconv.Itoa(config.Ports.ui));
}