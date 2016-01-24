#!/usr/bin/env ruby

require 'aws-sdk'
require 'net/ssh'
require 'net/scp'
require 'pry'
require 'droplet_kit'
require 'awesome_print'
require 'securerandom'
require 'digest'

# ------------------------------------------------------------------------------
#   Credentials
# ------------------------------------------------------------------------------
# pick up AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY by default from
# environment
Aws.config.update({
  region: 'us-west-2',
})

DO_TOKEN = ENV['DO_TOKEN_ONLINEJUDGE']
# ------------------------------------------------------------------------------

require 'trollop'
opts = Trollop::options do
  opt :list_run_instances, "List running 'run' instances"
  opt :list_loadbalancer_instances, "List running 'loadbalancer' instances"
  opt :list_run_images, "List 'run' images"
  opt :list_loadbalancer_images, "List 'loadbalancer' images"
  opt :get_latest_run_image, "Get latest run image for a region", :type => :string
  opt :get_latest_loadbalancer_image, "Get latest loadbalancer image for a region", :type => :string
  opt :destroy_image, "BE CAREFUL Delete an image by ID BE CAREFUL", :type => :string
  opt :start_instance_with_latest_run_image, "Start instance with latest run image in region", :type => :string
  opt :start_instance_with_latest_loadbalancer_image, "Start instance with latest loadbalancer image in region", :type => :string
  opt :refresh_dns, "Update DNS A records and health checks in Route 53"
  opt :refresh_loadbalancers, "Update loadbalancer backend hosts"
  opt :remove_dns, "Remove a region from DNS A records and health checks in Route 53", :type => :string
  opt :power_cycle_droplet, "Power cycle a droplet by ID (use list-instances first)", :type => :string
  opt :destroy_droplet, "BE CAREFUL Delete an instance by ID (use list-instances first) BE CAREFUL", :type => :string
end

def get_droplet_ipv4_address(droplet)
  ipv4 = droplet.networks.v4
  if ipv4.length == 0
    nil
  else
    ipv4[0].ip_address
  end
end

def get_droplet_hash(droplet)
  return {
    "droplet_id" => droplet.id,
    "name" => droplet.name,
    "created_at" => droplet.created_at,
    "status" => droplet.status,
    "created_at" => droplet.created_at,
    "region_slug" => droplet.region.slug,
    "image_id" => droplet.image.id,
    "image_name" => droplet.image.name,
    "ipv4_address" => get_droplet_ipv4_address(droplet),
  }
end

def get_instances(do_client, type)
  droplets = do_client.droplets.all
  droplets.map { |d|
    get_droplet_hash(d)
  }.select { |d|
    d['image_name'].include? "rsc #{type}"
  }.sort_by { |d|
    [d['region_slug'], d['image_name'], d['ipv4_address']]
  }
end

def get_images(do_client, type)
  images = do_client.images.all
  images.select { |i|
    i.name.include? "rsc #{type}"
  }.sort_by { |i|
    i.name
  }
end

def get_latest_image(do_client, region, type)
  get_images(do_client, type).select { |i|
    i.name.include? "rsc #{type} #{region}"
  }.sort_by { |i|
    i.name
  }[-1]
end

def generate_instance_name(region, type)
  slug = Array.new(8){rand(36).to_s(36)}.join
  "#{type}.#{slug}.#{region}.runsomecode.com"
end

def wait_for_droplet_status(do_client, droplet, status)
  success = false
  start = Time.now.to_i
  timeout = 120
  droplet_id = droplet["droplet_id"]
  while 1
    droplet = do_client.droplets.find(id: droplet_id)
    if droplet["status"] == status
      success = true
      break
    end
    puts "."
    sleep(5)
    break if (Time.now.to_i - start) >= timeout
  end
  ap droplet
  if not success
    puts "failed to start within timeout"
  else
    puts "successfully started within timeout"
  end
end

def start_instance(do_client, image, region_slug, type, ssh_key_name="Mill", size_slug="512mb")
  ssh_key = do_client.ssh_keys.all.select { |s| s.name == ssh_key_name }[0]
  size = do_client.sizes.all.select { |s| s.slug == size_slug }[0]
  region = do_client.regions.all.select { |r| r.slug == region_slug }[0].slug
  instance_name = generate_instance_name(region, type)
  droplet_params = DropletKit::Droplet.new(
    name: instance_name,
    size: size_slug,
    image: image.id,
    region: region,
    ssh_keys: [ssh_key.id]
  )
  droplet = do_client.droplets.create(droplet_params)
  droplet = do_client.droplets.find(id: droplet.id)
  droplet = get_droplet_hash(droplet)
  ap droplet
  wait_for_droplet_status(do_client, droplet, "active")
end

def refresh_dns(do_client)
  instances = get_instances(do_client, "loadbalancer")
  route53 = Aws::Route53::Client.new()
  update_health_checks(route53, instances)
  update_hosted_zones(route53, instances)
end

def remove_dns(do_client, region_to_remove)
  instances = get_instances(do_client, "loadbalancer").select { |i|
    i['region_slug'] != region_to_remove
  }
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
  instances_by_ip = Hash[instances.map { |i| [i["ipv4_address"], i] }]
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
    if not rrset_ip_map.include? i["ipv4_address"]
      puts "need to create rrset for: "
      ap i, :sort_keys => true
      health_check = health_checks.select{ |h| h.health_check_config.ip_address == i["ipv4_address"] }[0]
      changes << {
        action: "CREATE",
        resource_record_set: {
          name: "backend.runsomecode.com.",
          type: "A",
          set_identifier: i["name"],
          region: digital_ocean_to_ec2[i["region_slug"]],
          ttl: 60,
          resource_records: [
            value: i["ipv4_address"],
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
  instances_by_ip = Hash[instances.map { |i| [i["ipv4_address"], i] }]
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
    if not health_checks_by_ip.include? i["ipv4_address"]
      puts "creating health check for:"
      ap i, :sort_keys => true
      resp = route53.create_health_check({
        caller_reference: SecureRandom.uuid(),
        health_check_config: {
          ip_address: i["ipv4_address"],
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

def power_cycle_droplet(do_client, droplet_id)
  ap do_client.droplet_actions.power_cycle(id: droplet_id), :sort_keys => true
end

def destroy_droplet(do_client, droplet_id)
  ap do_client.droplets.delete(id: droplet_id), :sort_keys => true
end

def destroy_image(do_client, image_id)
  # TODO how?
  ap Digitalocean::Image.destroy(image_id), :sort_keys => true
end

def refresh_loadbalancers(do_client)
  get_instances(do_client, "loadbalancer").each { |lb|
    lines_to_insert = get_loadbalancer_lines_to_insert(do_client, lb)
    refresh_loadbalancer(lb, lines_to_insert)
  }
end

def get_loadbalancer_lines_to_insert(do_client, lb)
  run_instances = get_instances(do_client, "run").sort_by { |d|
    [d["region_slug"], d["image_name"], d["ipv4_address"]]
  }
  same_region = run_instances.select { |d| d["region_slug"] == lb["region_slug"] }
  different_region = run_instances.select { |d| d["region_slug"] != lb["region_slug"] }
  run_instances = same_region + different_region
  run_instances.map { |d|
    name = d["name"]
    ipv4_address = d["ipv4_address"]
    "    server #{name} #{ipv4_address}:80 maxconn 1 check\n"
  }
end

def refresh_loadbalancer(droplet, lines_to_insert)
  puts "refresh_loadbalancer entry for droplet: "
  ap droplet

  keys = ['/Users/ai/.ssh/digitalocean']
  remote_path = "/etc/haproxy/haproxy.cfg"
  Net::SCP.start(droplet["ipv4_address"], "root", :keys => keys) do |scp|
    data_before = scp.download!(remote_path)
    lines = data_before.lines.dup
    start_line = lines.index "# --- server block start ---\n"
    end_line = lines.index "# --- server block end ---\n"
    lines[start_line+1,end_line-start_line-1] = lines_to_insert
    data_after = lines.join("")
    if Digest::SHA256.hexdigest(data_before) == Digest::SHA256.hexdigest(data_after)
      puts "No change in config file, won't update."
      return
    end 
    puts "Change in config file, will update."
    puts "Before:"
    puts data_before
    puts "After:"
    puts data_after
    scp.upload! StringIO.new(data_after), remote_path
  end

  Net::SSH.start(droplet["ipv4_address"], "root", :keys => keys) do |ssh|
    ssh.exec!("/etc/init.d/haproxyctl reload") do |channel, stream, data|
      ap data
    end
  end
end

if __FILE__ == $0
  do_client = DropletKit::Client.new(access_token: DO_TOKEN)
  if opts[:list_run_instances]
    instances = get_instances(do_client, "run")
    ap instances, :sort_keys => true
  elsif opts[:list_loadbalancer_instances]
    instances = get_instances(do_client, "loadbalancer")
    ap instances, :sort_keys => true
  elsif opts[:list_run_images]
    images = get_images(do_client, "run")
    ap images, :sort_keys => true
  elsif opts[:list_loadbalancer_images]
    images = get_images(do_client, "loadbalancer")
    ap images, :sort_keys => true
  elsif opts[:get_latest_run_image]
    region = opts[:get_latest_run_image]
    latest_image = get_latest_image(do_client, region, "run")
    ap latest_image, :sort_keys => true
  elsif opts[:get_latest_loadbalancer_image]
    region = opts[:get_latest_loadbalancer_image]
    latest_image = get_latest_image(do_client, region, "loadbalancer")
    ap latest_image, :sort_keys => true
  elsif opts[:start_instance_with_latest_run_image]
    region = opts[:start_instance_with_latest_run_image]
    latest_image = get_latest_image(do_client, region, "run")
    start_instance(do_client, latest_image, region, "run")
  elsif opts[:start_instance_with_latest_loadbalancer_image]
    region = opts[:start_instance_with_latest_loadbalancer_image]
    latest_image = get_latest_image(do_client, region, "loadbalancer")
    start_instance(do_client, latest_image, region, "loadbalancer")
  elsif opts[:destroy_image]
    destroy_image(do_client, opts[:destroy_image])
  elsif opts[:refresh_dns]
    refresh_dns(do_client)
  elsif opts[:refresh_loadbalancers]
    refresh_loadbalancers(do_client)
  elsif opts[:remove_dns]
    region = opts[:remove_dns]
    remove_dns(do_client, region)
  elsif opts[:power_cycle_droplet]
    power_cycle_droplet(do_client, opts[:power_cycle_droplet])
  elsif opts[:destroy_droplet]
    destroy_droplet(do_client, opts[:destroy_droplet])
  end
end