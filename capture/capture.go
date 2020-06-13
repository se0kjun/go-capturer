package capture

/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"io"
	"os"
	"sync"
	"syscall"
)

// CaptureFlag has stdout and stderr flag
type CaptureFlag struct {
	CaptureStdout bool
	CaptureStderr bool
}

// CaptureCgo captures stdout and stderr with cgo
func CaptureCgo(call func(), cflag CaptureFlag) (output []byte, err error) {
	if !cflag.CaptureStdout && !cflag.CaptureStderr {
		return nil, nil
	}

	var lockStdFileDescriptorsSwapping sync.Mutex
	var originalStdout, originalStderr, tmpFd int
	var e error
	lockStdFileDescriptorsSwapping.Lock()

	if originalStdout, e = syscall.Dup(syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()
		return nil, e
	}
	if originalStderr, e = syscall.Dup(syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()
		return nil, e
	}
	lockStdFileDescriptorsSwapping.Unlock()

	defer func() {
		lockStdFileDescriptorsSwapping.Lock()

		if e = syscall.Dup2(originalStdout, syscall.Stdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e = syscall.Close(originalStdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e = syscall.Dup2(originalStderr, syscall.Stderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e = syscall.Close(originalStderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}

		lockStdFileDescriptorsSwapping.Unlock()
	}()

	_, trash, _ := os.Pipe()
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer func() {
		e = r.Close()
		if e != nil {
			err = e
		}
		if w != nil {
			e = w.Close()
			if err != nil {
				err = e
			}
		}
	}()

	lockStdFileDescriptorsSwapping.Lock()

	if cflag.CaptureStdout {
		tmpFd = int(w.Fd())
	} else {
		tmpFd = int(trash.Fd())
	}

	if e = syscall.Dup2(tmpFd, syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	if cflag.CaptureStderr {
		tmpFd = int(w.Fd())
	} else {
		tmpFd = int(trash.Fd())
	}

	if e = syscall.Dup2(tmpFd, syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	out := make(chan []byte)
	go func() {
		defer func() {
			// If there is a panic in the function call, copying from "r" does not work anymore.
			_ = recover()
		}()

		var b bytes.Buffer

		_, err := io.Copy(&b, r)
		if err != nil {
			panic(err)
		}

		out <- b.Bytes()
	}()

	call()

	lockStdFileDescriptorsSwapping.Lock()

	C.fflush(C.stdout)
	C.fflush(C.stderr)

	err = w.Close()
	if err != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, err
	}
	w = nil

	if e = syscall.Close(syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}
	if e = syscall.Close(syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	return <-out, err
}
