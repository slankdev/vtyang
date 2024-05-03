#!/bin/sh

if [ $# -ne 2 ]; then
	echo "invalid command syntax" 1>&2
	echo "Usage: $0 <frr-path> <dst-dir:pkg/vtyang/testdata/yang/frr_mgmtd_minimal>" 1>&2
	echo "Example:" 1>&2
  echo " $0 ~/git/frr.10.0 ./pkg/vtyang/testdata/yang/frr_mgmtd_minimal" 1>&2
	exit 1
fi

FILES="
frr-filter.yang
frr-interface.yang
frr-nexthop.yang
frr-ripd.yang
frr-ripngd.yang
frr-route-types.yang
frr-routing.yang
frr-staticd.yang
frr-vrf.yang
frr-zebra.yang
frr-route-map.yang
frr-affinity-map.yang
frr-bfdd.yang
frr-if-rmap.yang
"

set -xe
for file in $FILES; do
  cp $1/yang/$file $2
done
