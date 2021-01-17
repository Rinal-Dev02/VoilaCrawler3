package main

import (
	"crypto/md5"
	"fmt"
	"os"
)

var (
	hostname string
	nodeId   string
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
	nodeId = fmt.Sprintf("%x", md5.Sum([]byte(hostname)))
}

func Hostname() string {
	return hostname
}

func NodeId() string {
	return nodeId
}
