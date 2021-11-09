package api

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/howeyc/fsnotify"
	"github.com/rs/zerolog/log"
)

// DynamicFile .
type DynamicFile struct {
	File       string
	UpdateFunc func(fileData []byte)
}

// LogFuncs .
type LogFuncs struct {
	Info  func(args ...interface{})
	Error func(args ...interface{})
}

// FilePathWatcher .
type FilePathWatcher struct {
	DynamicFiles []DynamicFile
	path         string
}

//NewFilePathWatcher .
func NewFilePathWatcher(path string, df []DynamicFile) *FilePathWatcher {
	return &FilePathWatcher{
		df,
		path,
	}
}

// Watch watches for changes to files in the given directory.
// This function is blocking so it is meant to be used by running `go Watch("path", done)`
// The watcher
func (w *FilePathWatcher) Watch(done chan bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Err(err).Msg("oops...")
	}

	defer watcher.Close()

	// Process events until watcher closes
	go func(fsw *fsnotify.Watcher) {
		for {
			select {
			case ev := <-fsw.Event:
				if ev.IsCreate() || ev.IsModify() {
					if err := w.UpdateDynamicFile(ev.Name); err != nil {
						log.Err(err).Msg("oops...")
					}
				}

			case err := <-fsw.Error:
				// watcher.Close() signals a nil error
				if err == nil {
					return
				}
				// continue watching
				log.Err(err).Msg("oops...")
			}
		}
	}(watcher)

	if err := watcher.Watch(w.path); err != nil {
		return err
	}

	// continuous loop until parent routine closes
	for {
		if _, ok := <-done; !ok {
			return nil
		}
	}
}

// UpdateDynamicFile sets the instance values for Secrets *secrets
func (w *FilePathWatcher) UpdateDynamicFile(file string) error {
	for _, df := range w.DynamicFiles {
		if strings.HasSuffix(file, df.File) {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return err
			}

			b, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}

			log.Info().Msgf("updating file: %s", file)

			df.UpdateFunc(b)

			return nil
		}
	}

	return nil
}
