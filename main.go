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

var (
	watcher   *fsnotify.Watcher
	xmutex    sync.Mutex
	cmdname   string
	arguments []string
)

func main() {

	// build cmd args
	if len(os.Args) >= 1 {
		cmdname = os.Args[1]
	}
	for i := 2; i < len(os.Args); i++ {
		arguments = append(arguments, os.Args[i])
	}

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project,
	// walk each file/directory searching for
	// directories
	pwd, _ := os.Getwd()
	go watchPeriodically(pwd, 10)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		timestamp := time.Now().Unix()
		xmutex.Lock()
		for {
			xmutex.Unlock()
			select {
			// watch for events
			// case event := <-watcher.Events:
			// fmt.Printf("EVENT! %#v\n", event)
			case <-watcher.Events:
				if time.Now().Unix()-timestamp > 5 {
					execcmd(cmdname, arguments, &xmutex)
					timestamp = time.Now().Unix()
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

// watchPeriodically adds sub directories periodically to watch, with the help
// of fsnotify which maintains a directory map rather than slice.
func watchPeriodically(directory string, interval int) {
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		<-done
		if err := filepath.Walk(directory, watchDir); err != nil {
			fmt.Println(err)
		}
		go func() {
			done <- struct{}{}
		}()
	}
}

func execcmd(c string, cargs []string, lock *sync.Mutex) {
	lock.Lock()
	defer lock.Unlock()

	fmt.Printf("Executing %s %v\n", c, cargs)
	cmd := exec.Command(c, cargs...)
	r, _ := cmd.CombinedOutput()
	if len(r) > 0 {
		fmt.Printf("%s\n", r)
	}
	time.Sleep(1 * time.Second)
}
