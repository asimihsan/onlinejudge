#!/usr/bin/env bash

# lxc needs reboot to configure properly
sudo reboot
sudo ifconfig eth0 down
sleep 60
