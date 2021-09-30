#!/usr/bin/ruby

require "./lib/suse/connect/yast.rb"
require "./lib/suse/connect/errors.rb"

# temporary script for manual testing SUSE::Connect::YaST.announce_system
# yast-registration/src/lib/registration/registration.rb

# requires rubygem-ffi which can be got from packagehub
# SUSEConnect -p PackageHub/15.2/x86_64
# zypper in ruby2.5-rubygem-ffi

if ARGV.length < 1
  puts "usage: ./test.rb regcode"
  puts "with debug: SCCDEBUG=1 ./test.rb regcode"
  exit
end

regcode = ARGV[0]
settings = {token: regcode, email: "user@acme.org"}
distro_target = ""

begin
  login, password = SUSE::Connect::YaST.announce_system(settings, distro_target)
rescue SUSE::Connect::ApiError => e
  puts "Error: #{e.message} #{e.code}"
  exit 1
end

puts "login: #{login}"
puts "password: #{password}"
