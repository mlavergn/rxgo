package rx

import (
	"bytes"
	"io"
	"io/ioutil"
	oslog "log"
	"os"
)

// Version export
const Version = "0.33.0"

// DEBUG flag for runtime
const DEBUG = false

// stand-in for system logger
var log *oslog.Logger

// debug logger
var dlog *oslog.Logger
var dfilter *string

// null logger
var lognull *oslog.Logger

// Config export
func Config(debug bool) {
	if debug {
		log = oslog.New(os.Stderr, "RxGo ", oslog.Ltime|oslog.Lshortfile)
	} else {
		log = lognull
	}
}

func debugFilter(reader io.Reader, filter []byte) {
	go func() {
		if filter == nil {
			io.Copy(os.Stderr, reader)
			return
		}
		buffer := make([]byte, 1024)
		for {
			readLen, _ := reader.Read(buffer)
			if bytes.Contains(buffer, filter) {
				os.Stderr.Write(buffer[:readLen])
			}
		}
	}()
}

// Debug export
func Debug(debug bool, filter []byte) {
	if debug {
		if filter != nil {
			reader, writer := io.Pipe()
			debugFilter(reader, filter)
			dlog = oslog.New(writer, "DEBUG ", oslog.Ltime|oslog.Lshortfile)
		} else {
			dlog = oslog.New(os.Stderr, "DEBUG ", oslog.Ltime|oslog.Lshortfile)

		}
	} else {
		dlog = lognull
	}
}

func init() {
	lognull = oslog.New(ioutil.Discard, "", 0)
	Config(DEBUG)
	Debug(false, nil)
}
