package internal

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/howeyc/fsnotify"
)

var fileTypesToWatch = []string{
	".gop",
}

type Watcher struct {
	eventHandler func(string)
	watch        *fsnotify.Watcher
}

func NewWatcher(handler func(string)) (*Watcher, error) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := Watcher{
		eventHandler: handler,
		watch:        watch,
	}

	go func() {
		for {
			select {
			case e := <-w.watch.Event:
				if w.shallFireEvent(e) {
					w.handleEvent(e)
				}
			case err := <-w.watch.Error:
				log.Println("failed to watch file changes:", err)
			}
		}
	}()

	return &w, nil
}

func (w Watcher) Watch(dir string) error {
	return filepath.Walk(dir, w.processWalked)
}

func (w Watcher) handleEvent(e *fsnotify.FileEvent) {
	if e.IsCreate() {
		w.handleCreateEvent(e)
		return
	}

	if w.isFileConcerned(e.Name) {
		w.eventHandler(e.Name)
	}
}

func (w Watcher) handleCreateEvent(e *fsnotify.FileEvent) {
	info, err := os.Stat(e.Name)
	if err != nil {
		log.Printf("failed to stat %q: %v\n", e.Name, err)
		return
	}

	if info.IsDir() {
		if err := filepath.Walk(e.Name, w.processWalked); err != nil {
			log.Printf("failed to watch %q: %v\n", e.Name, err)
		}
	} else if w.isFileConcerned(e.Name) {
		fmt.Println("watching", e.Name)
		if err := w.watch.Watch(e.Name); err != nil {
			log.Printf("failed to watch %q: %v\n", e.Name, err)
		}
		w.eventHandler(e.Name)
	}
}

func (w Watcher) isFileConcerned(file string) bool {
	ext := filepath.Ext(file)

	for _, tp := range fileTypesToWatch {
		if tp == ext {
			return true
		}
	}

	return false
}

func (w Watcher) processWalked(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if w.isFileConcerned(path) {
		return w.watch.Watch(path)
	}

	return nil
}

func (w Watcher) shallFireEvent(e *fsnotify.FileEvent) bool {
	// why we put IsAttrib() at the first,
	// because modify on vim fires the attrib event and modify event.
	if e.IsAttrib() {
		return false
	}
	if e.IsCreate() {
		return true
	}
	if e.IsDelete() {
		return true
	}
	if e.IsModify() {
		return true
	}
	if e.IsRename() {
		return true
	}

	return false
}
