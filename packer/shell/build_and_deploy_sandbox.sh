#!/usr/bin/env bash

set -e
set -x

cp -r /tmp/sandbox /home/ubuntu/.local/share/lxc/ubase/rootfs/home/ubuntu/sandbox
chown -R 101000:101000 /home/ubuntu/.local/share/lxc/ubase/rootfs/home/ubuntu/sandbox

ssh ubuntu@localhost "lxc-start -n ubase -d"
ssh ubuntu@localhost "lxc-wait -n ubase -s RUNNING"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'cd /home/ubuntu/sandbox && make && make install'"
ssh ubuntu@localhost "lxc-attach -n ubase -- poweroff"

rm -rf /home/ubuntu/.local/share/lxc/ubase/rootfs/home/ubuntu/sandbox