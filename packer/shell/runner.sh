#!/bin/sh

set -e

# ubuntu user also needs to be able to ssh to itself
su - ubuntu -c 'mkdir -p ~/.ssh'
su - ubuntu -c 'rm -f ~/.ssh/known_hosts'
su - ubuntu -c 'touch ~/.ssh/known_hosts'
su - ubuntu -c 'chmod 644 ~/.ssh/known_hosts'
su - ubuntu -c 'ssh-keyscan -H localhost >> ~/.ssh/known_hosts'

# Java libraries copied to LXC container
sudo chown ubuntu:ubuntu /home/ubuntu/*.jar

sudo chmod a+x /usr/local/bin/runner.linux

sudo tee /usr/local/bin/runner_via_ssh >/dev/null <<"EOF"
#!/usr/bin/env bash
rm -rf /tmp/foo
mkdir -p /tmp/foo
chown -R ubuntu:ubuntu /tmp/foo
chmod 777 /tmp/foo

mkdir -p ~/.ssh
rm -f ~/.ssh/known_hosts
touch ~/.ssh/known_hosts
chmod 644 ~/.ssh/known_hosts
ssh-keyscan -H localhost >> ~/.ssh/known_hosts

exec ssh ubuntu@localhost /usr/local/bin/runner.linux
EOF
sudo chmod a+x /usr/local/bin/runner_via_ssh

sudo init-checkconf /etc/init/runner.conf
