#!/usr/bin/ruby

require "./lib/suse/connect/yast.rb"
require "./lib/suse/connect/errors.rb"

# temporary script for manual testing SUSE::Connect::YaST.announce_system
# yast-registration/src/lib/registration/registration.rb

# requires rubygem-ffi which can be got from packagehub
# SUSEConnect -p PackageHub/15.2/x86_64
# zypper in ruby2.5-rubygem-ffi

if ARGV.length < 2
  puts "usage: ./test.rb distro_target regcode"
  puts "or: SCCDEBUG=1 ./test.rb distro_target regcode"
  exit
end

distro_target = ARGV[0]
regcode = ARGV[1]
settings = {token: regcode, email: "user@acme.org"}

begin
  login, password = SUSE::Connect::YaST.announce_system(settings, distro_target)
rescue SUSE::Connect::ApiError => e
  puts "Error: #{e.message} #{e.code}"
  exit 1
end

puts "login: #{login}, password: #{password}"

begin
  SUSE::Connect::YaST.credentials()
rescue => e
  puts e.inspect
end

SUSE::Connect::YaST.create_credentials_file(login, password)

creds = SUSE::Connect::YaST.credentials()
puts "#{creds}"
