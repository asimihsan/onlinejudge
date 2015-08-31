#!/usr/bin/env ruby

require 'aws-sdk'
require 'net/ssh'
require 'pry'
require 'digitalocean'
require 'awesome_print'
require 'securerandom'

# ------------------------------------------------------------------------------
#   Credentials
# ------------------------------------------------------------------------------
# pick up AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY by default from
# environment
Aws.config.update({
  region: 'us-west-2',
})

Digitalocean.client_id  = ENV['DIGITAL_OCEAN_CLIENT_ID']
Digitalocean.api_key    = ENV['DIGITAL_OCEAN_API_KEY']
# ------------------------------------------------------------------------------

require 'trollop'
opts = Trollop::options do
  opt :list_instances, "List running instances"
  opt :list_images, "List images"
  opt :get_latest_image, "Get latest image for a region", :type => :string
  opt :start_instance_with_latest_image, "Start instance with latest image in region", :type => :string
  opt :refresh_dns, "Update DNS A records and health checks in Route 53"
end

def update_droplet_with_extra_info(droplet)
    droplet.region_slug = Digitalocean::Region.find(droplet.region_id).region.slug
    droplet.image_name = Digitalocean::Image.find(droplet.image_id).image.name
    droplet
end  

def get_instances()
  droplets = Digitalocean::Droplet.all
  droplets.droplets.map { |d|
    update_droplet_with_extra_info(d)
  }.sort_by { |d|
    [d.region_slug, d.image_name, d.ip_address]
  }
end

def get_images()
  images = Digitalocean::Image.all
  images.images.select { |i|
    i.name.include? "rsc run"
  }.sort_by { |i|
    i.name
  }
end

def get_latest_image(region)
  get_images().select { |i|
    i.name.include? "rsc run #{region}"
  }.sort_by { |i|
    i.name
  }[-1]
end

def generate_instance_name(region)
  slug = Array.new(8){rand(36).to_s(36)}.join
  "run.#{slug}.#{region.slug}.runsomecode.com"
end

def wait_for_droplet_status(droplet, status)
  success = false
  start = Time.now.to_i
  timeout = 120
  while 1
    droplet = Digitalocean::Droplet.find(droplet.id).droplet
    if droplet.status == status
      success = true
      break
    end
    puts "."
    sleep(5)
    break if (Time.now.to_i - start) >= timeout
  end
  ap update_droplet_with_extra_info(droplet)
  if not success
    puts "failed to start within timeout"
  else
    puts "successfully started within timeout"
  end
end

def start_instance(image, region_slug, ssh_key_name="Mill", size_slug="512mb")
  ssh_key = Digitalocean::SshKey.all.ssh_keys.select { |s| s.name == ssh_key_name }[0]
  size = Digitalocean::Size.all.sizes.select { |s| s.slug == size_slug }[0]
  region = Digitalocean::Region.find(region_slug).region
  instance_name = generate_instance_name(region)
  droplet = Digitalocean::Droplet.create(
    {name: instance_name, size_id: size.id, image_id: image.id,
     region_id: region.id, ssh_key_ids: [ssh_key.id]}).droplet
  droplet = Digitalocean::Droplet.find(droplet.id).droplet
  droplet = update_droplet_with_extra_info(droplet)
  ap droplet
  wait_for_droplet_status(droplet, "active")
end

def refresh_dns()
  instances = get_instances()
  route53 = Aws::Route53::Client.new()
  update_health_checks(route53, instances)
  update_hosted_zones(route53, instances)
end

def update_hosted_zones(route53, instances)
  hosted_zone = route53.list_hosted_zones_by_name(
      {dns_name: "runsomecode.com."}).hosted_zones[0]
  rrsets = route53.list_resource_record_sets(
    {hosted_zone_id: hosted_zone.id}).resource_record_sets
  rrsets = rrsets.select{ |rrset|
    rrset.type == "A" and rrset.alias_target.nil?
  }
  
  digital_ocean_to_ec2 = {
    "lon1" => "eu-west-1",
    "sfo1" => "us-west-1",
    "sgp1" => "ap-southeast-1",
  }

  changes = maybe_delete_rrset_for_old_instance(route53, rrsets, instances)
  if changes.length > 0
    resp = route53.change_resource_record_sets({
      hosted_zone_id: hosted_zone.id,
      change_batch: {
        changes: changes,
      },
    })
    ap resp
  end
  changes = maybe_create_rrset_for_new_instance(route53, rrsets, instances, digital_ocean_to_ec2)
  if changes.length > 0
    resp = route53.change_resource_record_sets({
      hosted_zone_id: hosted_zone.id,
      change_batch: {
        changes: changes,
      },
    })
    ap resp
  end
end

def maybe_delete_rrset_for_old_instance(route53, rrsets, instances)
  instances_by_ip = Hash[instances.map { |i| [i.ip_address, i] }]
  changes = []
  rrsets.each { |rrset|
    if not instances_by_ip.include? rrset.resource_records[0].value
      puts "need to delete rrset: "
      ap rrset
      changes << {
        action: "DELETE",
        resource_record_set: {
          name: rrset.name,
          type: rrset.type,
          set_identifier: rrset.set_identifier,
          region: rrset.region,
          ttl: rrset.ttl,
          resource_records: rrset.resource_records,
          health_check_id: rrset.health_check_id,
        },
      }
    end
  }
  changes
end

def maybe_create_rrset_for_new_instance(route53, rrsets, instances, digital_ocean_to_ec2)
  rrset_ip_map = Hash[rrsets.map { |rrset| [rrset.resource_records[0].value, rrset] }]
  changes = []
  health_checks = route53.list_health_checks().health_checks
  instances.each { |i|
    if not rrset_ip_map.include? i.ip_address
      puts "need to create rrset for: "
      ap i, :sort_keys => true
      health_check = health_checks.select{ |h| h.health_check_config.ip_address == i.ip_address }[0]
      changes << {
        action: "CREATE",
        resource_record_set: {
          name: "backend.runsomecode.com.",
          type: "A",
          set_identifier: i.name,
          region: digital_ocean_to_ec2[i.region_slug],
          ttl: 60,
          resource_records: [
            value: i.ip_address,
          ],
          health_check_id: health_check.id,
        },
      }
    end
  }
  changes
end

def update_health_checks(route53, instances)
  health_checks = route53.list_health_checks().health_checks
  health_checks.each { |h|
    maybe_delete_health_check_for_old_instance(route53, h, instances)
  }
  maybe_create_health_check_for_new_instance(route53, health_checks, instances)
end

def maybe_delete_health_check_for_old_instance(route53, health_check, instances)
  instances_by_ip = Hash[instances.map { |i| [i.ip_address, i] }]
  if not instances_by_ip.include? health_check.health_check_config.ip_address
    puts "deleting health check: "
    ap health_check, :sort_keys => true
    resp = route53.delete_health_check({
      health_check_id: health_check.id
    })
    ap resp, :sort_keys => true
  end
end

def maybe_create_health_check_for_new_instance(route53, health_checks, instances)
  health_checks_by_ip = Hash[health_checks.map { |h| [h.health_check_config.ip_address, h] }]
  instances.each { |i|
    if not health_checks_by_ip.include? i.ip_address
      i = update_droplet_with_extra_info(i)
      puts "creating health check for:"
      ap i, :sort_keys => true
      resp = route53.create_health_check({
        caller_reference: SecureRandom.uuid(),
        health_check_config: {
          ip_address: i.ip_address,
          port: 80,
          type: "HTTP",
          resource_path: "/ping",
          request_interval: 30,
          failure_threshold: 3,
        },
      })
      ap resp, :sort_keys => true
    end
  }
end

if __FILE__ == $0
  if opts[:list_instances]
    instances = get_instances()
    ap instances, :sort_keys => true
  elsif opts[:list_images]
    images = get_images()
    ap images, :sort_keys => true
  elsif opts[:get_latest_image]
    latest_image = get_latest_image(opts[:get_latest_image])
    ap latest_image, :sort_keys => true
  elsif opts[:start_instance_with_latest_image]
    region = opts[:start_instance_with_latest_image]
    latest_image = get_latest_image(region)
    start_instance(latest_image, region)
  elsif opts[:refresh_dns]
    refresh_dns()
  end
end