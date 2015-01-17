#!/bin/bash

set -e

# Unattended upgrade, especially for grub updates
sudo apt-get --assume-yes --quiet update
sudo apt-get --assume-yes --quiet upgrade
sudo DEBIAN_FRONTEND=noninteractive apt-get -y -o \
    Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
    dist-upgrade

# lxc
sudo add-apt-repository --yes ppa:ubuntu-lxc/stable
sudo apt-get --assume-yes --quiet update
sudo apt-get --assume-yes --quiet install lxc python3-lxc

# ntp
sudo apt-get --assume-yes --quiet install ntp

# sandboxing
sudo apt-get --assume-yes --quiet install seccomp libseccomp-dev libseccomp2 apparmor-profiles apparmor-utils
sudo apt-get --assume-yes --quiet install pkg-config libglib2.0 libglib2.0-dev linux-headers-$(uname -r)

# C/C++
sudo apt-get --assume-yes --quiet install build-essential

# Python
sudo apt-get --assume-yes --quiet install python python-dev

# Ruby
sudo apt-get --assume-yes --quiet install ruby ruby-dev

# Java
sudo apt-get --assume-yes --quiet install default-jdk

# Misc
sudo apt-get --assume-yes --quiet install git htop silversearcher-ag \
    sshpass coreutils

# /usr/share/dict/words
sudo apt-get --assume-yes --quiet install wamerican

# zsh
sudo apt-get --assume-yes --quiet install zsh
curl -L http://install.ohmyz.sh | sh
sed -i 's/ZSH_THEME=.*$/ZSH_THEME="bira"/' ~/.zshrc

sudo apt-get clean

# Generate SSH keypair
ssh-keygen -b 2048 -t rsa -f $HOME/.ssh/id_rsa -q -N ""
