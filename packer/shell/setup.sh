#!/bin/bash

set -e

# add mirrors to speed up apt-get upgrade
insert_line_at_top_of_file() {
    sudo sed -i -e "1i$1\\" $2
}
#insert_line_at_top_of_file 'deb mirror://mirrors.ubuntu.com/mirrors.txt trusty-security main restricted universe multiverse' '/etc/apt/sources.list'
#insert_line_at_top_of_file 'deb mirror://mirrors.ubuntu.com/mirrors.txt trusty-backports main restricted universe multiverse' '/etc/apt/sources.list'
#insert_line_at_top_of_file 'deb mirror://mirrors.ubuntu.com/mirrors.txt trusty-updates main restricted universe multiverse' '/etc/apt/sources.list'
#insert_line_at_top_of_file 'deb mirror://mirrors.ubuntu.com/mirrors.txt trusty main restricted universe multiverse' '/etc/apt/sources.list'

# Unattended upgrade, especially for grub updates
# http://askubuntu.com/questions/194651/why-use-apt-get-upgrade-instead-of-apt-get-dist-upgrade
sudo apt-get --assume-yes --quiet update
sudo DEBIAN_FRONTEND=noninteractive apt-get -y \
    -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
    dist-upgrade

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

sudo apt-get clean

# Generate SSH keypair
ssh-keygen -b 2048 -t rsa -f $HOME/.ssh/id_rsa -q -N ""
