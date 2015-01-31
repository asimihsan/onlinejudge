#!/bin/bash

# References:
# - http://www.flockport.com/lxc-using-unprivileged-containers/

set -e
set -x

# Need a new user for unprivileged containers, since on Digital Ocean we're root
sudo useradd -d /home/ubuntu -m ubuntu -p password -s /bin/zsh
sudo -u ubuntu tee /home/ubuntu/.zshrc >/dev/null <<"EOF"
# Set up the prompt
autoload -Uz promptinit
promptinit
prompt adam1

setopt histignorealldups sharehistory

# Use emacs keybindings even if our EDITOR is set to vi
bindkey -e

# Keep 1000 lines of history within the shell and save it to ~/.zsh_history:
HISTSIZE=1000
SAVEHIST=1000
HISTFILE=~/.zsh_history

# Use modern completion system
autoload -Uz compinit
compinit

zstyle ':completion:*' auto-description 'specify: %d'
zstyle ':completion:*' completer _expand _complete _correct _approximate
zstyle ':completion:*' format 'Completing %d'
zstyle ':completion:*' group-name ''
zstyle ':completion:*' menu select=2
eval "$(dircolors -b)"
zstyle ':completion:*:default' list-colors ${(s.:.)LS_COLORS}
zstyle ':completion:*' list-colors ''
zstyle ':completion:*' list-prompt %SAt %p: Hit TAB for more, or the character to insert%s
zstyle ':completion:*' matcher-list '' 'm:{a-z}={A-Z}' 'm:{a-zA-Z}={A-Za-z}' 'r:|[._-]=* r:|=* l:|=*'
zstyle ':completion:*' menu select=long
zstyle ':completion:*' select-prompt %SScrolling active: current selection at %p%s
zstyle ':completion:*' use-compctl false
zstyle ':completion:*' verbose true

zstyle ':completion:*:*:kill:*:processes' list-colors '=(#b) #([0-9]#)*=0=01;31'
zstyle ':completion:*:kill:*' command 'ps -u $USER -o pid,%cpu,tty,cputime,cmd'

ZSH_THEME="bira"
EOF

sudo apt-get --assume-yes --quiet install systemd-services uidmap

# Set up a cgroup environment for the new user
sudo cgm create all ubuntu
sudo cgm chown all ubuntu $(id -u ubuntu) $(id -g ubuntu)
cgm movepid all ubuntu $$

# User ID mappings for unprivileged containers. Maps root in container
# to unprivileged users in host, hence prevents accessing /proc and
# /sys files owner by root in host, and other root files.
sudo -u ubuntu mkdir -p /home/ubuntu/.local/share/lxcsnaps
sudo -u ubuntu mkdir -p /home/ubuntu/.local/share/lxc
sudo -u ubuntu mkdir -p /home/ubuntu/.cache/lxc
sudo -u ubuntu mkdir -p /home/ubuntu/.config/lxc
sudo -u ubuntu tee /home/ubuntu/.config/lxc/default.conf >/dev/null <<EOF
lxc.id_map = u 0 100000 65536
lxc.id_map = g 0 100000 65536

# Allow external network access via a bridge.
lxc.network.type = veth
lxc.network.link = lxcbr0
lxc.network.flags = up
lxc.network.hwaddr = 00:16:3e:xx:xx:xx

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

# Allocate additional uids and guids to username
sudo usermod --add-subuids 100000-165536 ubuntu
sudo usermod --add-subgids 100000-165536 ubuntu

# The last number indicates how many containers can share this
# interface.
echo "ubuntu veth lxcbr0 8" | sudo tee -a /etc/lxc/lxc-usernet

# Create a base container. After setting it up we'll stop it and use it
# as the basis for future clones of other containers. The only way to
# start/stop is to SSH in as the user (what an almighty hack!)
# https://lists.linuxcontainers.org/pipermail/lxc-users/2014-August/007485.html
sudo -u ubuntu mkdir -p /home/ubuntu/.ssh
sudo -u ubuntu touch /home/ubuntu/.ssh/authorized_keys
sudo -u ubuntu chmod 644 /home/ubuntu/.ssh/authorized_keys
cat /root/.ssh/id_rsa.pub >> /home/ubuntu/.ssh/authorized_keys

# Give root the host keys for the ubuntu user and itself
mkdir -p ~/.ssh
rm -f ~/.ssh/known_hosts
touch ~/.ssh/known_hosts
chmod 644 ~/.ssh/known_hosts
ssh-keyscan -H localhost >> ~/.ssh/known_hosts

ssh ubuntu@localhost "lxc-create -t download -n ubase -- -d ubuntu -r trusty -a amd64"
ssh ubuntu@localhost "lxc-start -n ubase -d"
ssh ubuntu@localhost "lxc-wait -n ubase -s RUNNING"
ssh ubuntu@localhost "cat /etc/resolv.conf | lxc-attach -n ubase -- tee /etc/resolv.conf >/dev/null"
sleep 10
ssh ubuntu@localhost "lxc-attach -n ubase -- ifconfig"
ssh ubuntu@localhost "lxc-attach -n ubase -- apt-get --quiet --assume-yes update"
ssh ubuntu@localhost "lxc-attach -n ubase -- apt-get --quiet --assume-yes upgrade"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'apt-get --quiet --assume-yes install \
    build-essential python python-dev ruby ruby-dev git nodejs npm \
    pkg-config libglib2.0 libglib2.0-dev linux-headers-$(uname -r) \
    openssh-server ufw coreutils seccomp libseccomp-dev libseccomp2 \
    wamerican libcap-dev strace software-properties-common \
    python-software-properties'"
ssh ubuntu@localhost "lxc-attach -n ubase -- add-apt-repository -y ppa:webupd8team/java"
ssh ubuntu@localhost "lxc-attach -n ubase -- apt-get --quiet --assume-yes update"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'echo debconf shared/accepted-oracle-license-v1-1 select true | sudo debconf-set-selections'"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'echo debconf shared/accepted-oracle-license-v1-1 seen true | sudo debconf-set-selections'"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'apt-get --quiet --assume-yes install \
    oracle-java7-installer'"
ssh ubuntu@localhost "lxc-attach -n ubase -- update-java-alternatives -s java-7-oracle"
ssh ubuntu@localhost "lxc-attach -n ubase -- apt-get --quiet --assume-yes install oracle-java7-set-default"
ssh ubuntu@localhost "lxc-attach -n ubase -- apt-get clean"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'id -u ubuntu &>/dev/null && userdel ubuntu'"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c '[ -d /home/ubuntu ] && rm -rf /home/ubuntu'"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'useradd -m -s /bin/bash ubuntu'"
ssh ubuntu@localhost "lxc-attach -n ubase -- bash -c 'yes password | passwd ubuntu'"
ssh ubuntu@localhost "lxc-stop -n ubase"
ssh ubuntu@localhost "lxc-wait -n ubase -s STOPPED"

# Clone some more containers using snapshots. Since we're using a
# directory backing store this enables a copy-on-write overlap
# filesystem.
# lxc-clone -s -o ubase -n u1
