#!/bin/bash

set -e

# User ID mappings for unprivileged containers. Maps root in container
# to unprivileged users in host, hence prevents accessing /proc and
# /sys files owner by root in host, and other root files.
mkdir -p ~/.config/lxc
tee ~/.config/lxc/default.conf >/dev/null <<EOF
lxc.id_map = u 0 100000 65536
lxc.id_map = g 0 100000 65536

# Allow external network access via a bridge.
lxc.network.type = veth
lxc.network.link = lxcbr0

# Resource limits
# https://www.kernel.org/doc/Documentation/cgroups/
# https://docs.oracle.com/cd/E37670_01/E37355/html/ol_cgroups.html
lxc.cgroup.memory.limit_in_bytes = 400M
lxc.cgroup.cpu.shares = 100
lxc.cgroup.cpuset.cpus = 0

# /usr/include/capability.h
lxc.cap.drop = audit_control
lxc.cap.drop = audit_write
lxc.cap.drop = mac_admin
lxc.cap.drop = mac_override
lxc.cap.drop = mknod
# lxc.cap.drop = net_admin  # needed to set up container
lxc.cap.drop = setfcap
lxc.cap.drop = setpcap
lxc.cap.drop = sys_admin
lxc.cap.drop = sys_boot
lxc.cap.drop = sys_module
lxc.cap.drop = sys_nice
lxc.cap.drop = sys_pacct
lxc.cap.drop = sys_rawio
lxc.cap.drop = sys_resource
lxc.cap.drop = sys_time
lxc.cap.drop = sys_tty_config
EOF

# The last number indicates how many containers can share this
# interface.
echo "$USER veth lxcbr0 8" | sudo tee -a /etc/lxc/lxc-usernet

# Create a base container. After setting it up we'll stop it and use it
# as the basis for future clones of other containers.
lxc-create -t download -n ubase -- -d ubuntu -r trusty -a amd64
lxc-start -n ubase -d
lxc-wait -n ubase -s RUNNING
cat /etc/resolv.conf | lxc-attach -n ubase -- tee /etc/resolv.conf >/dev/null
sleep 10
lxc-attach -n ubase -- ifconfig
lxc-attach -n ubase -- apt-get --quiet --assume-yes update
lxc-attach -n ubase -- apt-get --quiet --assume-yes upgrade
lxc-attach -n ubase -- bash -c 'apt-get --quiet --assume-yes install \
    build-essential python python-dev ruby ruby-dev default-jdk \
    pkg-config libglib2.0 libglib2.0-dev linux-headers-$(uname -r) \
    openssh-server ufw coreutils seccomp libseccomp-dev libseccomp2 \
    wamerican libcap-dev strace'
lxc-attach -n ubase -- apt-get clean
lxc-attach -n ubase -- bash -c 'id -u ubuntu &>/dev/null && userdel ubuntu'
lxc-attach -n ubase -- bash -c '[ -d /home/ubuntu ] && rm -rf /home/ubuntu'
lxc-attach -n ubase -- bash -c 'useradd -m -s /bin/bash ubuntu'
lxc-attach -n ubase -- bash -c 'yes password | passwd ubuntu'

lxc-stop -n ubase

# Clone some more containers using snapshots. Since we're using a
# directory backing store this enables a copy-on-write overlap
# filesystem.
# lxc-clone -s -o ubase -n u1
