#!/bin/sh

set -e

sudo chmod a+x /usr/local/bin/user_data.linux
sudo init-checkconf /etc/init/user_data.conf
