#!/bin/bash

set -e

# Unattended upgrade, especially for grub updates
sudo apt-get --assume-yes --quiet update
sudo apt-get --assume-yes --quiet upgrade

# lxc
sudo add-apt-repository --yes ppa:ubuntu-lxc/stable
sudo apt-get --assume-yes --quiet update
sudo apt-get --assume-yes --quiet install lxc python3-lxc

# ntp
sudo apt-get --assume-yes --quiet install ntp

# Misc
sudo apt-get --assume-yes --quiet install git htop silversearcher-ag \
    sshpass coreutils build-essential

# /usr/share/dict/words
sudo apt-get --assume-yes --quiet install wamerican

# zsh
sudo apt-get --assume-yes --quiet install zsh
curl -L http://install.ohmyz.sh | zsh
sed -i 's/ZSH_THEME=.*$/ZSH_THEME="bira"/' ~/.zshrc

# Get Gandi and intermediate CA chain for golang HTTP clients
(wget --no-check-certificate -q -O - https://www.gandi.net/static/CAs/GandiStandardSSLCA2.pem && \
    wget -q -O - http://crt.usertrust.com/USERTrustRSAAddTrustCA.crt | \
    openssl x509 -inform der -outform pem) > /etc/ssl/certs/GandiStandardSSLCA2.pem

sudo apt-get clean

# Generate SSH keypair
ssh-keygen -b 2048 -t rsa -f $HOME/.ssh/id_rsa -q -N ""
