package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/calavera/dkvolume"
)

type volume struct {
	name        string
	connections int
}

type glusterfsDriver struct {
	root    string
	servers []string
	volumes map[string]*volume
	m       sync.Mutex
}

func newGlusterfsDriver(root string, servers []string) glusterfsDriver {
	return glusterfsDriver{
		root:    root,
		servers: servers,
		volumes: map[string]*volume{}}
}

func (d glusterfsDriver) Create(r dkvolume.Request) dkvolume.Response {
	log.Printf("Creating volume %s\n", r.Name)
	return dkvolume.Response{}
}

func (d glusterfsDriver) Remove(r dkvolume.Request) dkvolume.Response {
	log.Printf("Removing volume %s\n", r.Name)
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)

	if s, ok := d.volumes[m]; ok {
		if s.connections <= 1 {
			delete(d.volumes, m)
		}
	}
	return dkvolume.Response{}
}

func (d glusterfsDriver) Path(r dkvolume.Request) dkvolume.Response {
	return dkvolume.Response{Mountpoint: d.mountpoint(r.Name)}
}

func (d glusterfsDriver) Mount(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Mounting volume %s on %s\n", r.Name, m)

	s, ok := d.volumes[m]
	if ok && s.connections > 0 {
		s.connections++
		return dkvolume.Response{Mountpoint: m}
	}

	fi, err := os.Lstat(m)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m, 0755); err != nil {
			return dkvolume.Response{Err: err.Error()}
		}
	} else if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	if fi != nil && !fi.IsDir() {
		return dkvolume.Response{Err: fmt.Sprintf("%v already exist and it's not a directory", m)}
	}

	if err := d.mountVolume(r.Name, m); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	d.volumes[m] = &volume{name: r.Name, connections: 1}

	return dkvolume.Response{Mountpoint: m}
}

func (d glusterfsDriver) Unmount(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Unmounting volume %s from %s\n", r.Name, m)

	if s, ok := d.volumes[m]; ok {
		if s.connections == 1 {
			if err := d.unmountVolume(m); err != nil {
				return dkvolume.Response{Err: err.Error()}
			}
		}
		s.connections--
	} else {
		return dkvolume.Response{Err: fmt.Sprintf("Unable to find volume mounted on %s", m)}
	}

	return dkvolume.Response{}
}

func (d *glusterfsDriver) mountpoint(name string) string {
	return filepath.Join(d.root, name)
}

func (d *glusterfsDriver) mountVolume(name, destination string) error {
	server := d.servers[rand.Intn(len(d.servers))]

	cmd := fmt.Sprintf("glusterfs --log-level=DEBUG --volfile-id=%s --volfile-server=%s %s", name, server, destination)
	if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
		log.Println(string(out))
		return err
	}
	return nil
}

func (d *glusterfsDriver) unmountVolume(target string) error {
	cmd := fmt.Sprintf("umount %s", target)
	if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
		log.Println(string(out))
		return err
	}
	return nil
}
