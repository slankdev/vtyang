# DEMO: FRR-mgmtd unix-socket api

## What's Northbound-API

gRPC + Yang model baseのユーザインターフェース.
Yang backendに対応したFRR daemon (isisdなど)で使うことができる.
FRR全体に対して一つのAPIを持つのではなく, 現在はdaemonごとに個別にgRPC server
が動作するので, Cisco RouterをgRPCで設定したい. 的なユースケースには
足りないものがたくさんある.
最近はmgmtdというそのような背景と課題を解決するものができてきたので
ゆくゆくはそっちに切り替わる認識をしている.

- https://docs.frrouting.org/en/latest/grpc.html
- https://docs.frrouting.org/projects/dev-guide/en/latest/northbound/northbound.html

## Snippet

```
log file /tmp/frr.log
debug northbound
!
router isis 1
 net 10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00
 flex-algo 128
  advertise-definition
  priority 100
 !
!

grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
-d '{"type":0,"encoding":0,"with_defaults":true,"path":["/"]}' \
localhost:9001 frr.Northbound.Get

grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
-d '{"type":0,"encoding":0,"with_defaults":true,"path":["/frr-interface:lib"]}' \
localhost:9001 frr.Northbound.Get | jq .data.data -r | less

grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
-d '{"type":1,"encoding":0,"with_defaults":true,"path":["/frr-isisd:isis"]}' \
localhost:9001 frr.Northbound.Get

grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
-d '{"type":1,"encoding":0,"with_defaults":false,"path":["/isis/instance/flex-algos"]}' localhost:9001 \
frr.Northbound.Get | jq .data.data -r

grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
localhost:9001 frr.Northbound.CreateCandidate | jq

grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
-d '{"candidateId":2,"delete":[{"path":"/isis/instance/flex-algos"}]}' \
localhost:9001 frr.Northbound.EditCandidate | jq

grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
-d '{"candidate_id":2,"phase":4}' \
localhost:9001 frr.Northbound.Commit | jq

cat /tmp/data.json
{
  "candidate_id": 4,
  "type": 0,
  "config": {
    "data": "{\"frr-isisd:flex-algos\":{\"flex-algo\":[{\"flex-algo\":128,\"advertise-definition\":true,\"priority\":100}]}}"
  }
}

cat /tmp/data.json | grpcurl -plaintext -import-path ~/git/frr/grpc \
-proto frr-northbound.proto -d @ localhost:9001 frr.Northbound.LoadToCandidate | jq
{}

./configure \
    --prefix=/usr \
    --includedir=\${prefix}/include \
    --bindir=\${prefix}/bin \
    --sbindir=\${prefix}/lib/frr \
    --libdir=\${prefix}/lib/frr \
    --libexecdir=\${prefix}/lib/frr \
    --sysconfdir=/etc \
    --localstatedir=/var \
    --with-moduledir=\${prefix}/lib/frr/modules \
    --enable-configfile-mask=0640 \
    --enable-logfile-mask=0640 \
    --enable-snmp=agentx \
    --enable-multipath=64 \
    --enable-user=frr \
    --enable-group=frr \
    --enable-vty-group=frrvty \
    --with-pkg-git-version \
    --enable-grpc \
    --with-pkg-extra-version=-MyOwnFRRVersion
```

```
$ ls ~/
~/git/frr
~/git/vtyang

$ source ~/git/vtyang/cmd/frr-agent/cli.bash
$ frr-get "/frr-isisd:isis" | jq .data.data -r | jq
{}
$ frr-create-candidate-config
{
  "candidateId": 1
}
$ frr-load-to-candidate-config 1 ~/git/vtyang/cmd/frr-agent/example-isis-config.json
{}
$ frr-commit 2 4
$ frr-get "/frr-isisd:isis" | jq .data.data -r | jq
{
  "frr-isisd:isis": {
    "instance": [
      {
        "area-tag": "1",
        "vrf": "default",
        "area-address": [
          "10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00"
        ]
      }
    ]
  }
}
```


```
set isis instance 1 default area-address 10.0000.0000.0000.0000.0000.0000.0000.0000.0000.00
commit
```

```
delete isis
commit
```

```
set isis instance 1 default segment-routing enabled true
set isis instance 1 default segment-routing prefix-sid-map prefix-sid 1.1.1.1/32 sid-value 1
set isis instance 1 default flex-algos flex-algo 200 advertise-definition true
set isis instance 1 default flex-algos flex-algo 200 priority 100
```

```
do show mgmt get-config /

/frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']/frr-staticd:staticd/route-list[prefix='1.1.1.1/32'][afi-safi='frr-routing:ipv4-unicast']/prefix 1.1.1.1/32
```

```
configure
log file /tmp/frr.log
debug mgmt frontend

! zebra? multi daemon?
ip prefix-list hoge seq 10 permit 10.255.0.0/16 le 32
ip prefix-list hoge seq 15 permit 10.254.0.0/16 le 32

! staticd
ip route 1.1.1.1/32 Null0
ip route 2.2.2.2/32 Null0

! zebra
interface dum0
 description dum0-interface-comment
 ip address 10.255.10.1/24
 exit
interface dum1
 description dum1-interface-comment
 ip address 10.255.11.1/24
 exit

exit
```

```
do show mgmt get-config /
do configure terminal file-lock

! zebra? multi daemon?
mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/action permit
mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix 10.255.0.0/16
mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='10']/ipv4-prefix-length-lesser-or-equal 32
mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='15']/action permit
mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='15']/ipv4-prefix 10.254.0.0/16
mgmt set-config /frr-filter:lib/prefix-list[type='ipv4'][name='hoge']/entry[sequence='15']/ipv4-prefix-length-lesser-or-equal 32

! staticd
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default'] {}
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']/frr-staticd:staticd/route-list[prefix='1.1.1.1/32'][afi-safi='frr-routing:ipv4-unicast'] {}
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']/frr-staticd:staticd/route-list[prefix='1.1.1.1/32'][afi-safi='frr-routing:ipv4-unicast']/path-list[table-id='0'][distance='1'] {}
mgmt set-config /frr-routing:routing/control-plane-protocols/control-plane-protocol[type='frr-staticd:staticd'][name='staticd'][vrf='default']/frr-staticd:staticd/route-list[prefix='1.1.1.1/32'][afi-safi='frr-routing:ipv4-unicast']/path-list[table-id='0'][distance='1']/frr-nexthops/nexthop[nh-type='blackhole'][vrf='default'][gateway=''][interface='(null)'] {}

! zebra
mgmt set-config /frr-interface:lib/interface[name='dum0'] {}
mgmt set-config /frr-interface:lib/interface[name='dum0']/description dum0-interface-comment
mgmt set-config /frr-interface:lib/interface[name='dum0']/frr-zebra:zebra/ipv4-addrs[ip='10.255.10.1'][prefix-length='24'] {}
mgmt set-config /frr-interface:lib/interface[name='dum1'] {}
mgmt set-config /frr-interface:lib/interface[name='dum1']/description dum1-interface-comment
mgmt set-config /frr-interface:lib/interface[name='dum1']/frr-zebra:zebra/ipv4-addrs[ip='10.255.11.1'][prefix-length='24'] {}

mgmt commit check
mgmt commit apply
do show mgmt datastore-contents json 
```

```
set lib prefix-list ipv4 hoge entry 10 action permit
set lib prefix-list ipv4 hoge entry 10 ipv4-prefix 10.255.0.0/16
set lib prefix-list ipv4 hoge entry 10 ipv4-prefix-length-lesser-or-equal 32
set lib prefix-list ipv4 hoge entry 15 action permit
set lib prefix-list ipv4 hoge entry 15 ipv4-prefix 10.254.0.0/16
set lib prefix-list ipv4 hoge entry 15 ipv4-prefix-length-lesser-or-equal 32

set routing control-plane-protocols control-plane-protocol staticd staticd default staticd route-list 1.1.1.1/32 ipv4-unicast path-list 0 1 frr-nexthops nexthop blackhole default "" ""

set lib interface dum0 description dum0-interface-comment
set lib interface dum0 zebra ipv4-addrs 10.255.10.1 24
set lib interface dum1 description dum1-interface-comment
set lib interface dum1 zebra ipv4-addrs 10.255.11.1 24

```

## References
- https://web.sfc.wide.ad.jp/~irino/blog/2023/04/02/frr-grpc/
