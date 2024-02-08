package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	watcher  *fsnotify.Watcher
	command  string
	interval = (int64)(2)
)

func main() {

	var err error

	for i := 1; i < len(os.Args); i++ {
		command += os.Args[i]
		command += ""
	}

	// creates a new file watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	// starting at the root of the project,
	// walk each file/directory searching for
	// directories
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	go watchPeriodically(pwd, 10)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		timestamp := time.Now().Unix()
		for {
			select {
			// watch for events
			// case event := <-watcher.Events:
			// fmt.Printf("EVENT! %#v\n", event)
			case <-watcher.Events:
				if time.Now().Unix()-timestamp > interval {
					execcmd(command)
					timestamp = time.Now().Unix()
				}

			// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
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
	printedOnce := false
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		<-done
		if err := filepath.Walk(directory, watchDir); err != nil {
			if !printedOnce {
				fmt.Fprintln(os.Stderr, "Error with filepath.Walk on", directory, err)
				printedOnce = true
			}
		}
		go func() {
			done <- struct{}{}
		}()
	}
}

func execcmd(c string) {

	time.Sleep(time.Duration(interval) * time.Second / 2)
	hr, min, sec := time.Now().Clock()

	fmt.Printf("%0d:%02d:%02d - %s\n", hr, min, sec, c)
	r, err := exec.Command("bash", "-c", c).CombinedOutput() // Why worry ?
	if err != nil {
		log.Print(err)
	}
	if err == nil && len(r) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", r)
	}
}
