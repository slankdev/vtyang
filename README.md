# vtyang
Yang based VTY

## Quick Start

```
curl -Lo vtyang https://github.com/slankdev/vtyang/releases/download/branch-main/pr-3.vtyang.linux-arm64.go1.17
chmod +x vtyang
./vtyang version
./vtyang agent --run-path /tmp/vtyang
```

```
docker run --rm docker pull ghcr.io/slankdev/vtyang-vtyang:pr-3 bash
vtyang version
vtyang agent --run-path /tmp/vtyang
```

## Deprecated Info
```
mkdir -p /tmp/vtyang/run
echo '{"users": {"user": [{"name": "hiroki"}]}}' > /tmp/config.json
./vtyang agent --dbpath /tmp/config.json --run-path /tmp/vtyang/run
```

## FRRouting Integration

expected frr config data
```
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

configure cli
```
```
