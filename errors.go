package main

import "errors"

var (
	errUsedHostInterface = errors.New("Container used Host interface")
	errNotFound          = errors.New("Container not found")
	errPIDFileNotFound   = errors.New("PID file not found")
)
