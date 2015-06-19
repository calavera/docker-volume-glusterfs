# -*- mode: ruby -*-
# vi: set ft=ruby :

user_config = """
adduser --disabled-password david
adduser david sudo
chown -R david /home/david/go

mkdir -p /home/david/.ssh
chown -R david /home/david
cat << END > /home/david/.ssh/authorized_keys
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDFoY7zn6ZP4EgovBHnVsqPeJ16LuRw4u0Yv8ScHGCMIRMTUM1vW+hO8VIH7DjgabhvzV/OJ4/BEFAJ8NYVouTsW89+vPHqJtWpMUqUN1iCGahYKwXgTXNuHCv+NUMc2rrHP+hizDc/s64djxdGT6iMNKHg9uLv7HLGQFjVSXmCK9Mrdg+d/H3Yhrsoqavdn61Y/H7CxMCvaGsnFIDPsI/BkG4p28GsNPyFpIZoPXdbBXwyaU6EGTPgQgpizbZ1HkMTKNYJeLQLP05Uwa/5KHLZAp74UVYfaSXTqsZrDtGZ8Q4pbKsQ11jrOj99vIDSs9el/9FT0pYaqEMPKbur/5wD david.calavera@gmail.com
END

cat << END > /etc/sudoers.d/david
david ALL=(ALL) NOPASSWD:ALL
END
"""

hosts_config = """cat << END >> /etc/hosts
172.21.12.11 gfs-server-1
172.21.12.12 gfs-server-2
172.21.12.13 gfs-server-3

172.21.12.10 gfs-client-1
172.21.12.20 gfs-client-2
END
"""

server_shell = """
DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -yq python-software-properties
DEBIAN_FRONTEND=noninteractive add-apt-repository ppa:semiosis/ubuntu-glusterfs-3.5
DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -yq glusterfs-server

#{user_config}

#{hosts_config}
"""

client_shell = %Q{
DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -yq python-software-properties
DEBIAN_FRONTEND=noninteractive add-apt-repository ppa:semiosis/ubuntu-glusterfs-3.5
DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -yq glusterfs-client

#{user_config}

cd /home/david
curl -z go1.4.2.linux-amd64.tar.gz -L -O https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
tar -C /usr/local -zxf /home/david/go1.4.2.linux-amd64.tar.gz

cat << END > /etc/profile.d/go.sh
export GOPATH=\\/home/david/go
export PATH=\\$GOPATH/bin:/usr/local/go/bin:\\$PATH
END

cat << END > /etc/sudoers.d/go
Defaults env_keep += "GOPATH"
END

#{hosts_config}
}

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/trusty64"
  # We setup three nodes to be gluster hosts, and two gluster client to mount the volume
  3.times do |i|
    id = i+1
    config.vm.define vm_name = "gfs-server-#{id}" do |config|
      config.vm.hostname = vm_name
      ip = "172.21.12.#{id+10}"
      config.vm.network :private_network, ip: ip
      config.vm.provision :shell, :inline => server_shell, :privileged => true
    end
  end

  2.times do |i|
    id = i+1
    config.vm.define vm_name = "gfs-client-#{id}" do |config|
      config.ssh.forward_agent = true
      config.vm.synced_folder ".", "/home/david/go/src/github.com/calavera/docker-volume-glusterfs", create: true

      config.vm.hostname = vm_name
      ip = "172.21.12.#{id * 10}"
      config.vm.network :private_network, ip: ip
      config.vm.provision :shell, :inline => client_shell, :privileged => true
    end
  end
end
