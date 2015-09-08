#!/usr/bin/env ruby

require 'aws-sdk'
require 'net/ssh'
require 'net/scp'
require 'pry'
require 'digitalocean'
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

Digitalocean.client_id  = ENV['DIGITAL_OCEAN_CLIENT_ID']
Digitalocean.api_key    = ENV['DIGITAL_OCEAN_API_KEY']
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
  opt :power_cycle_droplet, "Power cycle a droplet by ID (use list-instances first)", :type => :string
  opt :destroy_droplet, "BE CAREFUL Delete an instance by ID (use list-instances first) BE CAREFUL", :type => :string
end

def update_droplet_with_extra_info(droplet)
    droplet.region_slug = Digitalocean::Region.find(droplet.region_id).region.slug
    droplet.image_name = Digitalocean::Image.find(droplet.image_id).image.name
    droplet
end  

def get_instances(type)
  droplets = Digitalocean::Droplet.all
  droplets.droplets.map { |d|
    update_droplet_with_extra_info(d)
  }.select { |d|
    d.image_name.include? "rsc #{type}"
  }.sort_by { |d|
    [d.region_slug, d.image_name, d.ip_address]
  }
end

def get_images(type)
  images = Digitalocean::Image.all
  images.images.select { |i|
    i.name.include? "rsc #{type}"
  }.sort_by { |i|
    i.name
  }
end

def get_latest_image(region, type)
  get_images(type).select { |i|
    i.name.include? "rsc #{type} #{region}"
  }.sort_by { |i|
    i.name
  }[-1]
end

def generate_instance_name(region, type)
  slug = Array.new(8){rand(36).to_s(36)}.join
  "#{type}.#{slug}.#{region.slug}.runsomecode.com"
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

def start_instance(image, region_slug, type, ssh_key_name="Mill", size_slug="512mb")
  ssh_key = Digitalocean::SshKey.all.ssh_keys.select { |s| s.name == ssh_key_name }[0]
  size = Digitalocean::Size.all.sizes.select { |s| s.slug == size_slug }[0]
  region = Digitalocean::Region.find(region_slug).region
  instance_name = generate_instance_name(region, type)
  droplet = Digitalocean::Droplet.create(
    {name: instance_name, size_id: size.id, image_id: image.id,
     region_id: region.id, ssh_key_ids: [ssh_key.id]}).droplet
  droplet = Digitalocean::Droplet.find(droplet.id).droplet
  droplet = update_droplet_with_extra_info(droplet)
  ap droplet
  wait_for_droplet_status(droplet, "active")
end

def refresh_dns()
  instances = get_instances("loadbalancer")
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

def power_cycle_droplet(droplet_id)
  ap Digitalocean::Droplet.power_cycle(droplet_id), :sort_keys => true
end

def destroy_droplet(droplet_id)
  ap Digitalocean::Droplet.destroy(droplet_id), :sort_keys => true
end

def destroy_image(image_id)
  ap Digitalocean::Image.destroy(image_id), :sort_keys => true
end

def refresh_loadbalancers()
  get_instances("loadbalancer").each { |lb|
    lines_to_insert = get_loadbalancer_lines_to_insert(lb)
    refresh_loadbalancer(lb, lines_to_insert)
  }
end

def get_loadbalancer_lines_to_insert(lb)
  run_instances = get_instances("run").sort_by { |d|
    [d.region_slug, d.image_name, d.ip_address]
  }
  same_region = run_instances.select { |d| d.region_slug == lb.region_slug }
  different_region = run_instances.select { |d| d.region_slug != lb.region_slug }
  run_instances = same_region + different_region
  run_instances.map { |d|
    "    server #{d.name} #{d.ip_address}:80 check\n"
  }
end

def refresh_loadbalancer(droplet, lines_to_insert)
  puts "refresh_loadbalancer entry for droplet: "
  ap droplet

  keys = ['/Users/ai/.ssh/digitalocean']
  remote_path = "/etc/haproxy/haproxy.cfg"
  Net::SCP.start(droplet.ip_address, "root", :keys => keys) do |scp|
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
    scp.upload! StringIO.new(data_after), remote_path
  end

  Net::SSH.start(droplet.ip_address, "root", :keys => keys) do |ssh|
    ssh.exec!("/etc/init.d/haproxyctl reload") do |channel, stream, data|
      ap data
    end
  end
end

if __FILE__ == $0
  if opts[:list_run_instances]
    instances = get_instances("run")
    ap instances, :sort_keys => true
  elsif opts[:list_loadbalancer_instances]
    instances = get_instances("loadbalancer")
    ap instances, :sort_keys => true
  elsif opts[:list_run_images]
    images = get_images("run")
    ap images, :sort_keys => true
  elsif opts[:list_loadbalancer_images]
    images = get_images("loadbalancer")
    ap images, :sort_keys => true
  elsif opts[:get_latest_image]
    latest_image = get_latest_image(opts[:get_latest_image])
    ap latest_image, :sort_keys => true
  elsif opts[:start_instance_with_latest_run_image]
    region = opts[:start_instance_with_latest_run_image]
    latest_image = get_latest_image(region, "run")
    start_instance(latest_image, region, "run")
  elsif opts[:start_instance_with_latest_loadbalancer_image]
    region = opts[:start_instance_with_latest_loadbalancer_image]
    latest_image = get_latest_image(region, "loadbalancer")
    start_instance(latest_image, region, "loadbalancer")
  elsif opts[:destroy_image]
    destroy_image(opts[:destroy_image])
  elsif opts[:refresh_dns]
    refresh_dns()
  elsif opts[:refresh_loadbalancers]
    refresh_loadbalancers()
  elsif opts[:power_cycle_droplet]
    power_cycle_droplet(opts[:power_cycle_droplet])
  elsif opts[:destroy_droplet]
    destroy_droplet(opts[:destroy_droplet])
  end
end