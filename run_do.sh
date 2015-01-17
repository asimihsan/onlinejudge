#!/usr/bin/env bash

# List images
#curl -X GET -H 'Content-Type: application/json' \
#    -H "Authorization: Bearer ${DO_TOKEN}" \
#    "https://api.digitalocean.com/v2/images?page=1&per_page=50" | jq '.'

# List regions
#curl -X GET -H 'Content-Type: application/json' \
#    -H "Authorization: Bearer ${DO_TOKEN}" \
#    "https://api.digitalocean.com/v2/regions" | jq '.'

# List SSH keys
#curl -X GET -H 'Content-Type: application/json' \
#    -H "Authorization: Bearer ${DO_TOKEN}" \
#    "https://api.digitalocean.com/v2/account/keys"  | jq '.'

# Create new droplet

IMAGE=$(curl -X GET -H 'Content-Type: application/json' \
    -H "Authorization: Bearer ${DO_TOKEN}" \
    "https://api.digitalocean.com/v2/images?page=1&per_page=50" | \
    jq '.images[] | select(.name | contains("onlinejudge"))')
echo running: "${IMAGE}"
IMAGE_ID=$(echo "${IMAGE}" | jq '.id')
echo $IMAGE_ID
curl -X POST -H 'Content-Type: application/json' \
    -H "Authorization: Bearer ${DO_TOKEN}" -d \
    "{\"name\":\"run1.runsomecode.com\",\"region\":\"sfo1\",\"size\":\"1g\",\"ssh_keys\":[\"610664\"],\"image\":\"${IMAGE_ID}\"}" \
    "https://api.digitalocean.com/v2/droplets" | jq '.'

# List all droplets, get names and ip addresses
#curl -X GET -H 'Content-Type: application/json' \
#    -H "Authorization: Bearer ${DO_TOKEN}" \
#    "https://api.digitalocean.com/v2/droplets?page=1&per_page=50" \
#    | jq '[.droplets[] | .name, .id, [.networks.v4[].ip_address]]'

# Delete a droplet
#curl -X DELETE -H 'Content-Type: application/json' \
#    -H "Authorization: Bearer ${DO_TOKEN}" \
#    "https://api.digitalocean.com/v2/droplets/3641247" 
