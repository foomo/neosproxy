#!/usr/bin/env ruby

log = ARGF.read

formatted = log.gsub(/commit ([\da-f]{40})\nAuthor: .*\nDate: +.*\n\n {4}(.*)\n(?:\ {4}.*\n)*/, '|\1|\2|')

puts formatted