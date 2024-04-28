# https://github.com/FRRouting/frr/blob/master/grpc/frr-northbound.proto
API=192.168.64.3:9001
API=localhost:9001

function frr-get {
  if [ $# -ne 1 ]; then
    echo "invalid command syntax" 1>&2
    echo "Usage: $0 <xpath>" 1>&2
    return 1
  fi

  # grpcurl -plaintext -import-path ~/git/frr/grpc -proto frr-northbound.proto \
  # -d '{"type":1,"encoding":0,"with_defaults":true,"path":["/frr-isisd:isis"]}' \
  # localhost:9001 frr.Northbound.Get
  # grpcurl -plaintext -import-path ~/git/frr/grpc \
  #   -proto frr-northbound.proto \
  #   -d '{"type":0,"encoding":0,"with_defaults":true,"path":["/frr-isisd:isis"]}' \
  #   $API frr.Northbound.Get
  grpcurl -plaintext -import-path ~/git/frr/grpc \
    -proto frr-northbound.proto \
    -d '{"type":1,"encoding":0,"with_defaults":false,"path":["'$1'"]}' \
    $API frr.Northbound.Get
}

function frr-create-candidate-config {
  grpcurl -plaintext -import-path ~/git/frr/grpc \
    -proto frr-northbound.proto \
    $API frr.Northbound.CreateCandidate | jq
}

function frr-load-to-candidate-config {
  if [ $# -ne 2 ]; then
    echo "invalid command syntax" 1>&2
    echo "Usage: $0 <candidateId> <config.data file>" 1>&2
    return 1
  fi

  ### EXAMPLE candidate file content
  # {
  #   "frr-isisd:flex-algos": {
  #     "flex-algo": [
  #       {
  #         "flex-algo": 128,
  #         "advertise-definition": true,
  #         "priority": 100
  #       }
  #     ]
  #   }
  # }
  cat <<EOF > /tmp/data.json
{
  "candidate_id": $1,
  "type": 0,
  "config": {
    "data": $(cat $2 | jq -c | jq -R)
  }
}
EOF

  cat /tmp/data.json | grpcurl -plaintext \
    -import-path ~/git/frr/grpc \
    -proto frr-northbound.proto \
    -d @ $API \
    frr.Northbound.LoadToCandidate | jq
}

function frr-commit {
  if [ $# -ne 2 ]; then
    echo "invalid command syntax" 1>&2
    echo "Usage: $0 <candidateId> <phase validate:0 prepare:1 abort:2 apply:3 all:4>" 1>&2
    return 1
  fi

  grpcurl -plaintext \
    -import-path ~/git/frr/grpc \
    -proto frr-northbound.proto \
    -d '{"candidate_id":'$1',"phase":'$2'}' \
    $API frr.Northbound.Commit | jq
}
