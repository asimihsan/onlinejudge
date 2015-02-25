#!/bin/sh

set -e

sudo chmod a+x /usr/local/bin/evaluator.linux
sudo init-checkconf /etc/init/evaluator.conf
