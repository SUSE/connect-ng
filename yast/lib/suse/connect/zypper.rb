require 'ostruct'

module SUSE
  module Connect
    module Zypper
      # this class is only needed to pass product to PackageSearch.search()
      # it could be removed if YaST switches to use generic OpenStruct for these calls
      class Product < OpenStruct
        def initialize(product_hash)
          product_hash[:identifier] = product_hash[:name]
          super
        end
      end
    end
  end
end
