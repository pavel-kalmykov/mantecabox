package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/phayes/permbits"
	"github.com/radovskyb/watcher"
)

func monitorChanges(w *watcher.Watcher) {
	for {
		select {
		case event := <-w.Event:
			fmt.Println(event) // Print the event's info.
			fileInfos, err := ioutil.ReadDir(event.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			for _, fileInfo := range fileInfos {
				fmt.Println("File:",
					permbits.FileMode(fileInfo.Mode()).String(),
					fileInfo.ModTime().Format(time.RFC822),
					fileInfo.Name())
			}
		case err := <-w.Error:
			fmt.Fprintln(os.Stderr, err)
		case <-w.Closed:
			return
		}
	}
}

func StartDaemon() {
	w := watcher.New()
	w.SetMaxEvents(1)
	go monitorChanges(w)
	dir, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	dir += "/Mantecabox/"
	if err := w.Add(dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Printf("Daemon started. Syncing in %v...\n", dir)
	if err := w.Start(time.Millisecond * 1000); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
