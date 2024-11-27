#!/bin/sh

set -e

# setting an address for loopback
ifconfig lo 127.0.0.1 mtu 9001
ifconfig

# adding a default route
ip route add default dev lo src 127.0.0.1
route -n

# iptables rules to route traffic to transparent proxy
iptables -A OUTPUT -t nat -p tcp --dport 1:65535 ! -d 127.0.0.1  -j DNAT --to-destination 127.0.0.1:1200
iptables -L -t nat

# generate identity key
/app/keygen-ed25519 --secret /app/id.sec --public /app/id.pub

# your custom setup goes here

# starting supervisord
cat /etc/supervisord.conf
/app/supervisord
