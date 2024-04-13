# DEMO: FRR-Agent for vtyang

## Quick Start

```
```

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
```
```
set isis instance 1 default segment-routing enabled true
set isis instance 1 default segment-routing prefix-sid-map prefix-sid 1.1.1.1/32 sid-value 1
set isis instance 1 default flex-algos flex-algo 200 advertise-definition true
set isis instance 1 default flex-algos flex-algo 200 priority 100
```

## References
- https://web.sfc.wide.ad.jp/~irino/blog/2023/04/02/frr-grpc/
