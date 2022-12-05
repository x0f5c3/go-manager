package config

import (
	"errors"
	"os"
	"sync"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"

	"github.com/x0f5c3/go-manager/internal/fsutil"
)

type WaitGroup struct {
	toTrack bool
	pb      *pterm.ProgressbarPrinter
	t       filesTryTracker
	*sync.Mutex
	*sync.WaitGroup
}

func (w *WaitGroup) Go(f func() configReadStatus) {
	w.Add(1)
	go func() {
		if w.toTrack {
			defer w.pb.Increment()
		}
		defer w.Done()
		w.Lock()
		defer w.Unlock()
		res := f()
		w.t = append(w.t, res)
	}()
}


type multiError []error

func (m multiError) Error() string {
	var errStr string
	for _, err := range m {
		errStr += err.Error() + "\n"
	}
	return errStr
}

func (w *WaitGroup) Wait() {

}

type configReadStatus struct {
	path        string
	exists      bool
	readSuccess bool
	err         error
}

type configReadSuccess struct {
	id []byte
}

type configReadTry struct {
	path        string
	exists      bool
	readSuccess bool
	err         error
}

func existFail(path string) configReadStatus {
	return configReadStatus{path: path, exists: false, err: errors.New("file does not exist"), readSuccess: false}
}

func readFail(path string, exists bool, err error) configReadStatus {
	return configReadStatus{path: path, exists: exists, err: err, readSuccess: false}
}

func readSuccess(path string) configReadStatus {
	return configReadStatus{path: path, exists: true, readSuccess: true, err: nil}
}

func tryReadOne(path string) configReadStatus {
	exists := fsutil.CheckExists(path)
	if !exists {
		return existFail(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return readFail(path, exists, err)
	}
	defer f.Close()
	err = viper.ReadConfig(f)
	if err != nil {
		return readFail(path, exists, err)
	}
	return readSuccess(path)
}

type filesTryTracker []configReadStatus

