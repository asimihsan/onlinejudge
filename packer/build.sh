#!/bin/bash

set -e

export PACKER_LOG=1

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR/../runner/ && make all-linux
cd $DIR/../evaluator/ && make all-linux
cd $DIR/../user_data/ && make all-linux
cd $DIR

BUILDER=$1
DIGITAL_OCEAN_REGION=$2

if [ "${BUILDER}" == "ebs" ]; then
    packer build --only=amazon-ebs packer.json

    # EC2 build often fails because the AMI takes a long time to change status
    # from pending to available. Recommend using the AWS Ruby SDK to
    # see if the failure is for this reason then wait ~10 minutes for the
    # status to change.

elif [ "${BUILDER}" == "digital_ocean" ]; then
    # second argument is region, e.g. sfo1, sgp1, lon1
    packer build --only=digitalocean \
        -var digital_ocean_region=${DIGITAL_OCEAN_REGION} \
        packer.json 
fi
