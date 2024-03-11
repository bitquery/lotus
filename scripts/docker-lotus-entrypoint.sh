#!/usr/bin/env bash

GATE="$LOTUS_PATH"/date_initialized

if [ ! -z $DOCKER_LOTUS_IMPORT_SNAPSHOT ] && ! $BACKFILL_MODE ; then
	# Don't init if already initialized.
	if [ ! -f "$GATE" ]; then
		echo importing minimal snapshot
		/usr/local/bin/lotus daemon --import-snapshot "$DOCKER_LOTUS_IMPORT_SNAPSHOT" --halt-after-import
		# Block future inits
		date > "$GATE"
	fi
fi

# import wallet, if provided
if [ ! -z $DOCKER_LOTUS_IMPORT_WALLET ]; then
	/usr/local/bin/lotus-shed keyinfo import "$DOCKER_LOTUS_IMPORT_WALLET"
fi


if  $BACKFILL_MODE && [ ! -f "$GATE" ]; then
     
  todoarr=()

  echo "rangestart is  $RANGESTART"
  echo "rangeend is $RANGEEND"

  for litelstitem in `curl https://forest-archive.chainsafe.dev/list/mainnet/lite | htmlq --attribute href a`
  do
    left=${litelstitem//*height_/}
    height=${left//.forest*/}
    lites[$height]=$litelstitem
  done

  for difflstitem in `curl https://forest-archive.chainsafe.dev/list/mainnet/diff | htmlq --attribute href a`
  do
    left=${difflstitem//*height_/}
    height=${left//.forest*/}
    diffs[$height]=$difflstitem
  done

  let liteindex=$RANGESTART/30000*30000 

  lite=("${lites[$liteindex]}")

  diffindex=$liteindex

  while [ $diffindex -lt $RANGEEND ]
  do
   todoarr+=("${diffs[$diffindex]}")
   let diffindex=$diffindex+3000
  done

  todoarr+=("https://forest-archive.chainsafe.dev/latest/mainnet/")

  echo "=================================== first processing lite snap  $lite"
    mkdir /var/lib/lotus/process
    cd /var/lib/lotus/process
    wget --retry-connrefused --waitretry=5 --read-timeout=20 --timeout=15 -t 0 -q -c $lite -O processfile.car.zst
    zstd -d processfile.car.zst
    rm -rf processfile.car.zst
    /usr/local/bin/lotus  daemon --halt-after-import --import-snapshot processfile.car 

     rm -rf processfile

  for value in "${todoarr[@]}"
  do
    echo "==================================    Processing $value"
    echo
    wget --retry-connrefused --waitretry=5 --read-timeout=20 --timeout=15 -t 0 -q -c $value -O processfile.car.zst
    zstd -d processfile.car.zst
    rm -rf processfile.car.zst
    /usr/local/bin/lotus-shed import-car processfile.car
    rm -rf processfile.car
    echo "------------------------- done $value"
    echo
    echo
    sleep 3 
  done

  date > "$GATE" 

fi

cd /

exec /usr/local/bin/lotus $@
