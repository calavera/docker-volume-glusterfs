package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/calavera/docker-volume-api"
)

var (
	serversList = flag.String("servers", "", "List of glusterfs servers")
	root        = flag.String("root", volumeapi.DefaultDockerRootDirectory, "Docker volumes root directory")
)

func main() {
	var Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(*serversList) == 0 {
		Usage()
		os.Exit(1)
	}

	servers := strings.Split(*serversList, ":")

	d := newGlusterfsDriver(*root, servers)
	h := volumeapi.NewVolumeHandler(d)
	fmt.Println("listening on :7878")
	fmt.Println(h.ListenAndServe("tcp", ":7878", ""))
}
