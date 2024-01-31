package main

import (
	"fmt"
	"os"
	"testing"
)

func Test_listFd(t *testing.T) {
	listFd(2244038)

	link, err := os.Readlink("/proc/2244038/fd/1")
	fmt.Printf("%+v, %+v", link, err)
}
