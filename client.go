package main

import (
	"fmt"
	"os"
	"path/filepath"
	"os/exec"
	"bytes"
	"log"
	"flag"
	"strings"

	"github.com/go-fsnotify/fsnotify"
	"net"
	"strconv"
	"bufio"
	"encoding/json"
)

var watcher *fsnotify.Watcher

func main() {
	libraryPath := flag.String("path", "./library/", "The folder of your Docker projects")
	updatesServer := flag.String("server", "192.168.33.1", "The server to send updates to")
	updatesPort := flag.Int("port", 9001, "The port to send updates to")

	flag.Parse()

	address := *updatesServer + ":" + strconv.Itoa(*updatesPort)

	log.Printf("Watching: %v", *libraryPath)
	log.Printf("Sending updates to: %v", address)

	watcher, _ = fsnotify.NewWatcher();
	defer watcher.Close()

	if err := filepath.Walk(*libraryPath, watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {

					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						fmt.Println(event.Name)
						runBuild(event, address)
					}
				}
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	<-done
}

func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}

func runBuild(event fsnotify.Event, address string) {
	conn, _ := net.Dial("tcp", address)

	// Build docker command
	cmdName := "docker"
	cmdArgs := []string{"build", "."}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = event.Name

	// Collect output
	out := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = out
	cmd.Stderr = stderr

	printCommand(cmd)

	// Run command
	err := cmd.Run()

	printError(err, stderr)
	printOutput(out.Bytes())

	// Send data to server
	type messageTemplate struct {
		Job string
		Message string
		Error string
	}

	message := messageTemplate{
		Job: event.Name,
		Message: string(out.Bytes()),
		Error: string(stderr.Bytes()),
	}

	b, err := json.Marshal(message)

	if err != nil {
		log.Fatal("error: ", err)
	}

	conn.Write(b)

	response, _ := bufio.NewReader(conn).ReadString('\n')

	fmt.Println(string(response))
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error, stderr *bytes.Buffer) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n %s\n", err.Error(), stderr.String()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}
