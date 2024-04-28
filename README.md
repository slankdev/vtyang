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

## Development Env Setup

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

PROTOC_VERSION=25.3
PROTOC_OS=linux
PROTOC_ARCH=aarch_64
mkdir -p /tmp/protoc
pushd /tmp/protoc
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-${PROTOC_OS}-${PROTOC_ARCH}.zip
unzip protoc-${PROTOC_VERSION}-${PROTOC_OS}-${PROTOC_ARCH}.zip
cp ./bin/protoc /usr/bin
protoc --version
```
