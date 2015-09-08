#!/bin/bash

set -e

export PACKER_LOG=1

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
INSTANCE_TYPE=$1
BUILDER=$2
DIGITAL_OCEAN_REGION=$3

if [ "${INSTANCE_TYPE}" == "run" ]; then
    PACKER_JSON=run.json
    cd $DIR/../runner/ && make all-linux
    cd $DIR/../evaluator/ && make all-linux
    cd $DIR/../user_data/ && make all-linux
    cd $DIR
elif [ "${INSTANCE_TYPE}" == "loadbalancer" ]; then
    PACKER_JSON=loadbalancer.json
fi

if [ "${BUILDER}" == "ebs" ]; then
    packer build --only=amazon-ebs "${PACKER_JSON}"

    # EC2 build often fails because the AMI takes a long time to change status
    # from pending to available. Recommend using the AWS Ruby SDK to
    # see if the failure is for this reason then wait ~10 minutes for the
    # status to change.

elif [ "${BUILDER}" == "digital_ocean" ]; then
    # second argument is region, e.g. sfo1, sgp1, lon1
    packer build --only=digitalocean \
        -var digital_ocean_region=${DIGITAL_OCEAN_REGION} \
        "${PACKER_JSON}" 
fi
