# Docker volume plugin for GlusterFS

This plugin uses GlusterFS as distributed data storage for containers.

## Installation

Using go (until we get proper binaries):

```
$ go get github.com/calavera/docker-volume-glusterfs
```

## Usage

This plugin doesn't create volumes in your GlusterFS cluster yet, so you'll have to create them yourself first.

1 - Start the plugin using this command:

```
$ sudo docker-volume-glusterfs -servers gfs-1:gfs-2:gfs-3
```

We use the flag `-servers` to specify where to find the GlusterFS servers. The server names are separated by colon.

2 - Start your docker containers with the option `--volume-driver=glusterfs` and use the first part of `--volume` to specify the remote volume that you want to connect to:

```
$ sudo docker run --volume-driver glusterfs --volume datastore:/data alpine touch /data/helo
```

See this video for a slightly longer usage explanation:

https://youtu.be/SVtsT9WVujs


## LICENSE

MIT
