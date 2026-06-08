package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var interval = int64(2)

func main() {
	// Reconstruction propre de la commande
	if len(os.Args) < 2 {
		log.Fatal("Usage: go-onchange <command>")
	}
	command := strings.Join(os.Args[1:], " ")

	// Création du watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Lancement de la surveillance périodique
	go watchPeriodically(watcher, pwd, 10)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	// Boucle principale d'événements
	go func() {
		timestamp := time.Now().Unix()
		for {
			select {
			case <-watcher.Events:
				if time.Now().Unix()-timestamp > interval {
					execcmd(command)
					timestamp = time.Now().Unix()
				}
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	// fmt.Printf("Surveillance démarrée pour : %s\n", command)
	<-done
}

// watchDir utilise le watcher passé en paramètre
func watchDir(watcher *fsnotify.Watcher) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			// Ignore le dossier .git pour éviter trop d'événements
			if strings.Contains(path, ".git") {
				return nil
			}
			return watcher.Add(path)
		}
		return nil
	}
}

func watchPeriodically(watcher *fsnotify.Watcher, directory string, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Premier passage immédiat
	_ = filepath.Walk(directory, watchDir(watcher))

	for range ticker.C {
		_ = filepath.Walk(directory, watchDir(watcher))
	}
}

func execcmd(c string) {
	time.Sleep(time.Duration(interval) * time.Second / 2)
	hr, min, sec := time.Now().Clock()

	fmt.Printf("%0d:%02d:%02d - Exécution : %s\n", hr, min, sec, c)
	r, err := exec.Command("bash", "-c", c).CombinedOutput()
	if err != nil {
		log.Printf("Erreur d'exécution : %v", err)
	}
	if len(r) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", r)
	}
}
