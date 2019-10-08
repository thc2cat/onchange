package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

//
var (
	watcher   *fsnotify.Watcher
	xmutex    sync.Mutex
	debut     string
	arguments []string
)

// main
func main() {

	if len(os.Args) > 2 {
		debut = os.Args[1]
	}
	for i := 2; i < len(os.Args); i++ {
		arguments = append(arguments, os.Args[i])
	}

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project, walk each file/directory searching for
	// directories
	here, _ := os.Getwd()
	if err := filepath.Walk(here, watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	//
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	//
	go func() {
		var lastmove string

		lastmove = ""

		xmutex.Lock()

		for {
			xmutex.Unlock()
			select {
			// watch for events
			case event := <-watcher.Events:
				fmt.Printf("EVENT! %#v\n", event)
				if lastmove != event.String() {
					execcmd()
					lastmove = event.String()
				}

				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
			xmutex.Lock()
		}
	}()

	<-done
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}

func execcmd() {
	xmutex.Lock()
	defer xmutex.Unlock()

	cmd := exec.Command(debut, arguments...)
	r, _ := cmd.CombinedOutput()
	fmt.Printf("OUTPUT YES: %s", r)

	time.Sleep(1 * time.Second)
}
