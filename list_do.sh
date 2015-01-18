#!/usr/bin/env bash

curl -X GET -H 'Content-Type: application/json' \
    -H "Authorization: Bearer ${DO_TOKEN}" \
    "https://api.digitalocean.com/v2/droplets?page=1&per_page=50" \
    | jq '[.droplets[] | select(.name | contains("runsomecode.com")) | .name, .id, [.networks.v4[].ip_address]]'