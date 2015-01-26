#!/usr/bin/env bash

set -e
set -x

SCP='sshpass -p password scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'

copy_code() {
    n=0
    until [ $n -ge 60 ]; do
        eval "${SCP}" -r /tmp/sandbox ubuntu@"${IP_ADDRESS}":/home/ubuntu && break
        n=$[$n+1]
        sleep 1
    done
}

mkdir -p ~/.ssh
rm -f ~/.ssh/known_hosts
touch ~/.ssh/known_hosts
chmod 644 ~/.ssh/known_hosts
ssh-keyscan -H localhost >> ~/.ssh/known_hosts

ssh ubuntu@localhost "lxc-start -n ubase -d"
ssh ubuntu@localhost "lxc-wait -n ubase -s RUNNING"
sleep 10
ssh ubuntu@localhost "lxc-attach -n ubase -- ifconfig"
IP_ADDRESS=$(ssh ubuntu@localhost "lxc-info -i -n ubase" | awk '{print $2}')
copy_code
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'cd /home/ubuntu/sandbox && make clean && make && make install'"
ssh ubuntu@localhost "lxc-attach -n ubase -- poweroff"
