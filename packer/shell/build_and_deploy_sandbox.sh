#!/usr/bin/env bash

set -e

LXC_CONTAINER_NAME=ubase
SCP='sshpass -p password scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -q'

copy_code() {
    n=0
    until [ $n -ge 60 ]; do
        eval "${SCP}" -r /tmp/sandbox ubuntu@"${IP_ADDRESS}":/home/ubuntu && break
        n=$[$n+1]
        sleep 1
    done
}

lxc-start -n "${LXC_CONTAINER_NAME}" -d
lxc-wait -n "${LXC_CONTAINER_NAME}" -s RUNNING
sleep 10
lxc-attach -n "${LXC_CONTAINER_NAME}" -- ifconfig
IP_ADDRESS=$(lxc-info -i -n "${LXC_CONTAINER_NAME}" | awk '{print $2}')
copy_code
lxc-attach -n "${LXC_CONTAINER_NAME}" -- \
    bash -c 'cd /home/ubuntu/sandbox && make clean && make && make install'
lxc-attach -n "${LXC_CONTAINER_NAME}" -- poweroff
