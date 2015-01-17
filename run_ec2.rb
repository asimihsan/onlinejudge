#!/usr/bin/env ruby

require 'aws-sdk'
require 'net/ssh'
require 'pry'

AWS.config(
  access_key_id: ENV['AWS_ACCESS_KEY_ID'],  # onlinejudge-ec2
  secure_access_key: ENV['AWS_SECURE_ACCESS_KEY'], # onlinejudge-ec2
)

region_to_ami = {
  'us-west-2' => 'ami-97f4a6a7',
}

if __FILE__ == $0
  region = 'us-west-2'

  ec2 = AWS.ec2(region: region)
  key_pair = ec2.key_pairs['onlinejudge']
  key_pair_path = '/Users/ai/Documents/keys/aws/onlinejudge.pem'
  security_group = ec2.security_groups['sg-0f0c4d6a']  # launch-wizard-1
  instance = ec2.instances.create(
    :image_id => region_to_ami[region],
    :instance_type => 'm3.medium',
    :key_name => key_pair.name,
    :security_groups => [security_group])
  sleep 1 while instance.status == :pending
  exit 1 unless instance.status == :running
  puts "instance is running on hostname: #{instance.dns_name}"

  #begin
  #    Net::SSH.start(instance.dns_name, "ubuntu",
  #                   :key_data => [key_pair.private_key]) do |ssh|
  #      puts "uname -s output"
  #      puts ssh.exec!("uname -a")
  #    end
  #rescue SystemCallError, Timeout::Error => e
  #  # Port 22 may take a while to become available
  #  sleep 1
  #  retry
  #end
  #binding.pry

end
