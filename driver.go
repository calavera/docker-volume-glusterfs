package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/calavera/docker-volume-glusterfs/rest"
)

type volume_name struct {
	name        string
	connections int
}

type glusterfsDriver struct {
	root       string
	restClient *rest.Client
	servers    []string
	volumes    map[string]*volume_name
	m          *sync.Mutex
}

func newGlusterfsDriver(root, restAddress, gfsBase string, servers []string) glusterfsDriver {
	d := glusterfsDriver{
		root:    root,
		servers: servers,
		volumes: map[string]*volume_name{},
		m:       &sync.Mutex{},
	}
	if len(restAddress) > 0 {
		d.restClient = rest.NewClient(restAddress, gfsBase)
	}
	return d
}

func (d glusterfsDriver) Create(r volume.Request) volume.Response {
	log.Printf("Creating volume %s\n", r.Name)
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)

	if _, ok := d.volumes[m]; ok {
		return volume.Response{}
	}

	if d.restClient != nil {
		exist, err := d.restClient.VolumeExist(r.Name)
		if err != nil {
			return volume.Response{Err: err.Error()}
		}

		if !exist {
			if err := d.restClient.CreateVolume(r.Name, d.servers); err != nil {
				return volume.Response{Err: err.Error()}
			}
		}
	}
	return volume.Response{}
}

func (d glusterfsDriver) Remove(r dkvolume.Request) volume.Response {
	log.Printf("Removing volume %s\n", r.Name)
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)

	if s, ok := d.volumes[m]; ok {
		if s.connections <= 1 {
			if d.restClient != nil {
				if err := d.restClient.StopVolume(r.Name); err != nil {
					return volume.Response{Err: err.Error()}
				}
			}
			delete(d.volumes, m)
		}
	}
	return volume.Response{}
}

func (d glusterfsDriver) Path(r volume.Request) volume.Response {
	return volume.Response{Mountpoint: d.mountpoint(r.Name)}
}

func (d glusterfsDriver) Mount(r volume.Request) volume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Mounting volume %s on %s\n", r.Name, m)

	s, ok := d.volumes[m]
	if ok && s.connections > 0 {
		s.connections++
		return volume.Response{Mountpoint: m}
	}

	fi, err := os.Lstat(m)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m, 0755); err != nil {
			return volume.Response{Err: err.Error()}
		}
	} else if err != nil {
		return volume.Response{Err: err.Error()}
	}

	if fi != nil && !fi.IsDir() {
		return volume.Response{Err: fmt.Sprintf("%v already exist and it's not a directory", m)}
	}

	if err := d.mountVolume(r.Name, m); err != nil {
		return volume.Response{Err: err.Error()}
	}

	d.volumes[m] = &volume_name{name: r.Name, connections: 1}

	return volume.Response{Mountpoint: m}
}

func (d glusterfsDriver) Unmount(r volume.Request) volume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Unmounting volume %s from %s\n", r.Name, m)

	if s, ok := d.volumes[m]; ok {
		if s.connections == 1 {
			if err := d.unmountVolume(m); err != nil {
				return volume.Response{Err: err.Error()}
			}
		}
		s.connections--
	} else {
		return volume.Response{Err: fmt.Sprintf("Unable to find volume mounted on %s", m)}
	}

	return volume.Response{}
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
