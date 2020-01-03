package rx

import (
	"io/ioutil"
	oslog "log"
	"os"
)

// -----------------------------------------------------------------------------
// Config

// Version export
const Version = "0.10.0"

// DEBUG flag for development
const DEBUG = false

// stand-in for system logger
var log *oslog.Logger

// Config export
func Config(debug bool) {
	if debug {
		log = oslog.New(os.Stderr, "RxGo ", oslog.Ltime|oslog.Lshortfile)
	} else {
		log = oslog.New(ioutil.Discard, "", 0)
	}
}

func init() {
	Config(DEBUG)
}
