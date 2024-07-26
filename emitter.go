package main

import (
	"github.com/myzhan/goreplay-udp/output"
	"io"
	"time"
)

// Start initialize loop for sending data from inputs to outputs
func Start(stop chan int) {

	for _, in := range Plugins.Inputs {
		go CopyMulty(in, Plugins.Outputs...)
	}

	for {
		select {
		case <-stop:
			finalize()
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// CopyMulty copies from 1 reader to multiple writers
func CopyMulty(src PluginReader, writers ...PluginWriter) (err error) {
	for {
		msg, err := src.PluginRead()
		if err != nil {
			if err == output.ErrorStopped || err == io.EOF {
				return nil
			}
			return err
		}
		for _, dst := range writers {
			if _, err := dst.PluginWrite(msg); err != nil {
				return err
			}
		}
	}

	return err
}
