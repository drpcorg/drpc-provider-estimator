#!/bin/bash
if [ "$1" = "" ]
then
  echo "Usage: $0 <hostname:port of dshackle>"
  exit
fi
set -x;
echo "classic requests"
./drpc-provider-estimator -t $1 -o out/classic.csv -d 10 > out/classic.out
sleep 5
echo "nano requests"
./drpc-provider-estimator -t $1 -o out/nano.csv -d 10 -p profiles/nano.yaml > out/nano.out
sleep 5
echo "blocknumber requests"
./drpc-provider-estimator -t $1 -o out/blocknumber.csv -d 10 -p profiles/blocknumber.yaml > out/blocknumber.out
sleep 5
echo "getbalance requests"
./drpc-provider-estimator -t $1 -o out/getbalance.csv -d 10 -p profiles/getbalance.yaml > out/getbalance.out
