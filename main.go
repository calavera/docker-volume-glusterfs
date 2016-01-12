package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
)

const glusterfsID = "_glusterfs"

var (
	defaultDir  = filepath.Join(volume.DefaultDockerRootDirectory, glusterfsID)
	serversList = flag.String("servers", "", "List of glusterfs servers")
	restAddress = flag.String("rest", "", "URL to glusterfsrest api")
	gfsBase     = flag.String("gfs-base", "/mnt/gfs", "Base directory where volumes are created in the cluster")
	root        = flag.String("root", defaultDir, "GlusterFS volumes root directory")
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

	d := newGlusterfsDriver(*root, *restAddress, *gfsBase, servers)
	h := volume.NewHandler(d)
	fmt.Println(h.ServeUnix("root", "glusterfs"))
}
