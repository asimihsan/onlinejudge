#!/bin/sh

set -e

sudo chmod a+x /usr/local/bin/evaluator.linux

sudo tee /usr/local/bin/evaluator_via_ssh >/dev/null <<"EOF"
#!/usr/bin/env bash
exec ssh ubuntu@localhost /usr/local/bin/evaluator.linux
EOF
sudo chmod a+x /usr/local/bin/evaluator_via_ssh

sudo init-checkconf /etc/init/evaluator.conf
