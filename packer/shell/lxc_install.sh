#!/bin/bash

# lxc
sudo add-apt-repository --yes ppa:ubuntu-lxc/stable
sudo apt-get --assume-yes --quiet update
sudo apt-get --assume-yes --quiet install lxc python3-lxc
