#!/bin/sh

set -e

sudo chmod a+x /root/runner.linux
sudo init-checkconf /etc/init/runner.conf
