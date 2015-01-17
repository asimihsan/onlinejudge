#!/bin/bash

set -e

export PACKER_LOG=1

#packer build --only=amazon-ebs packer.json
packer build --only=digitalocean packer.json

# EC2 build often fails because the AMI takes a long time to change status
# from pending to available. Recommend using the AWS Ruby SDK to
# see if the failure is for this reason then wait ~10 minutes for the
# status to change.
