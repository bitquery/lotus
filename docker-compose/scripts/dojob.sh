#!/bin/bash

for fl in `cat ./list`
do
 echo "==================================    Processing $fl"
 echo
 wget -c https://forest-archive.chainsafe.dev/mainnet/diff/$fl.zst
sleep 5
 zstd -d ./$fl.zst
 sleep 5
 rm -rf ./$fl.zst
 ./lotus daemon --halt-after-import --import-snapshot $fl
 rm -rf ./$fl
 echo "------------------------- done $fl"
 echo
 echo
 sleep 3 
done
