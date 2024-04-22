#!/bin/bash
set -xe
cat <<EOF | sudo vtysh 
conf
ip prefix-list hoge seq 10 permit 10.255.0.0/16 le 32
ip prefix-list hoge seq 15 permit 10.254.0.0/16 le 32
ip route 1.1.1.1/32 Null0
ip route 2.2.2.2/32 Null0
interface dum0
 description dum0-interface-comment
 ip address 10.255.10.1/24
 exit
interface dum1
 description dum1-interface-comment
 ip address 10.255.11.1/24
 exit
EOF
