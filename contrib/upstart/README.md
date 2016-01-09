# Upstart script for docker-volume-glusterfs
## configure your glusterfs nodes
Similar to the docker service you can pass parameters thru a config file
```bash
vi etc/default/docker-volume-glusterfs
```

## copy the config and init script inside of your /etc/ folder
```bash
sudo cp ./etc/default/docker-volume-glusterfs /etc/default/docker-volume-glusterfs
sudo cp ./etc/init/docker-volume-glusterfs
```
## reload the upstart configuration
```bash
sudo initctl reload-configuration
```

## start the docker-volume-glusterfs
```bash
sudo service docker-volume-glusterfs start
sudo service docker-volume-glusterfs status
```

## check the logs
```bash
sudo tail -f /var/log/upstart/docker-volume-glusterfs.log
```
