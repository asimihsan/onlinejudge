#!/usr/bin/env ruby

require 'aws-sdk'
require 'net/ssh'
require 'pry'
require 'digitalocean'

# pick up AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY by default from
# environment
Aws.config.update({
  region: 'us-west-2',
})

Digitalocean.client_id  = ENV['DIGITAL_OCEAN_CLIENT_ID']
Digitalocean.api_key    = ENV['DIGITAL_OCEAN_API_KEY']

if __FILE__ == $0
    route53 = Aws::Route53::Client.new()
    health_checks = route53.list_health_checks()

    hosted_zones = route53.list_hosted_zones_by_name(
        {dns_name: "runsomecode.com."})
    hosted_zone_id = hosted_zones[0][0].id

    droplets = Digitalocean::Droplet.all

    pry.binding
end