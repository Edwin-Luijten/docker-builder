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
	"net"
	"strconv"
	"bufio"
	"encoding/json"
	"time"

	"github.com/go-fsnotify/fsnotify"
	rand2 "math/rand"
)

var watcher *fsnotify.Watcher
var rand *rand2.Rand
var autoTag bool
var autoPush bool

type BuilderConfig struct {
	Files []string
}

type messageTemplate struct {
	Identifier string
	Type       string
	Job        string
	Message    string
	Error      string
}

func main() {
	rand = rand2.New(rand2.NewSource(time.Now().UnixNano()))

	libraryPath := flag.String("path", "./library/", "The folder of your Docker projects")
	updatesServer := flag.String("server.address", "192.168.33.1", "The server to send updates to")
	updatesPort := flag.Int("server.port", 9001, "The port to send updates to")
	identifier := flag.String("id", randomString(10), "A name which identifies this agent")
	tag := flag.Bool("auto.tag", false, "Tag the image automatically")
	push := flag.Bool("auto.push", false, "Push the image automatically")

	flag.Parse()

	address := *updatesServer + ":" + strconv.Itoa(*updatesPort)

	autoTag = *tag
	autoPush = *push

	log.Printf("Watching: %v", *libraryPath)
	log.Printf("Sending updates to: %[1]v as %[2]v", address, *identifier)

	watcher, _ = fsnotify.NewWatcher();
	defer watcher.Close()

	if err := filepath.Walk(*libraryPath, watchDir); err != nil {
		log.Fatal("ERROR", err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// Only handle writes
				if event.Op&fsnotify.Write == fsnotify.Write {

					// Sleep a bit so that temp files are gone
					time.Sleep(time.Millisecond * 100)

					// Only handle file changes
					if info, err := os.Stat(event.Name); err == nil && !info.IsDir() {

						log.Println(event.Name + " changed")
						dir := strings.Replace(event.Name, info.Name(), "", -1)

						sendMessage(createMessage(*identifier, "build-start", dir, "", ""), address)

						runBuild(dir, address, *identifier)
					}
				}
			case err := <-watcher.Errors:
				log.Println("ERROR", err)
			}
		}
	}()

	<-done
}

func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

	result := make([]byte, strlen)

	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}

func runBuild(dir string, address string, identifier string) {
	runCommand("docker", []string{"build", "."}, dir)
	sendMessage(createMessage(identifier, "build-end", dir, string(out.Bytes()), string(stderr.Bytes())), address)

	if autoTag {
		runCommand("docker", []string{"tag", "id", "repo:tag"}, dir)
	}

	if autoPush {
		runCommand("docker", []string{"push", "repo:tag"}, dir)
	}
}

func runCommand(command string, arguments []string, workingDir string) {
	cmd := exec.Command(command, arguments...)
	cmd.Dir = workingDir

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
}

func sendMessage(message []byte, address string) {
	conn, _ := net.Dial("tcp", address)

	conn.Write(message)

	response, _ := bufio.NewReader(conn).ReadString('\n')

	log.Println(string(response))
}

func createMessage(identifier string, jobType string, job string, message string, error string) []byte {
	template := messageTemplate{
		Identifier: identifier,
		Type:       jobType,
		Job:        job,
		Message:    message,
		Error:      error,
	}

	m, err := json.Marshal(template)

	if err != nil {
		log.Fatal("error: ", err)
	}

	return m
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
