// packfs / packfs
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package main

import (
    "log"
    "os"
    "os/exec"

    "github.com/hshimamoto/packfs/fs"
)

func main() {
    if len(os.Args) < 3 {
	log.Fatal("Usage:\n\tpackfs packfile MOUNTPOINT")
    }
    if len(os.Args) == 3 || os.Args[3] != "daemon" {
	cmd := exec.Command(os.Args[0], os.Args[1], os.Args[2], "daemon")
	// start and don't care about daemon process
	err := cmd.Start()
	if err != nil {
	    log.Fatal(err)
	}
	return
    }
    pfs, err := packfs.NewPackFS(os.Args[1], os.Args[2])
    if err != nil {
	log.Fatalf("Mount fail: %v\n", err)
    }
    pfs.Serve()
}
