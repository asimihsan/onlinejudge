#!/usr/bin/env bash

# Firewall, security
sudo apt-get --assume-yes --quiet install ufw
sudo ufw disable
sudo bash -c 'yes yes | ufw reset'
sudo ufw allow ssh
sudo ufw limit ssh
sudo ufw allow http
sudo bash -c 'yes yes | ufw enable'
sudo ufw status verbose

# sysctl
sudo tee /etc/sysctl.conf >/dev/null <<EOF
# Protect ICMP attacks
net.ipv4.icmp_echo_ignore_broadcasts = 1

# Turn on protection for bad icmp error messages
net.ipv4.icmp_ignore_bogus_error_responses = 1

# Turn on syncookies for SYN flood attack protection
net.ipv4.tcp_syncookies = 1

# Log suspcicious packets, such as spoofed, source-routed, and redirect
net.ipv4.conf.all.log_martians = 1
net.ipv4.conf.default.log_martians = 1

# Turn on execshild
kernel.exec-shield = 1
kernel.randomize_va_space = 1
EOF
