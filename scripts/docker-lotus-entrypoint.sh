#!/usr/bin/env bash

if [ ! -z $DOCKER_LOTUS_IMPORT_SNAPSHOT ] && [ $BACKFILL_MODE -ne "true"  ]; then
	GATE="$LOTUS_PATH"/date_initialized
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

if [ $BACKFILL_MODE -eq "true" ]; then
     
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

  todoarr+=("${lites[$liteindex]}")

  diffindex=$liteindex

  while [ $diffindex -lt $RANGEEND ]
  do
   todoarr+=("${diffs[$diffindex]}")
   let diffindex=$diffindex+3000
  done

  todoarr+=("https://forest-archive.chainsafe.dev/latest/mainnet/")

  for value in "${todoarr[@]}"
  do
    echo "==================================    Processing $value"
    echo
    wget -c $value -O processfile.zst
    zstd -d processfile.zst
    rm -rf processfile.zst
    /usr/local/bin/lotus daemon --halt-after-import --import-snapshot processfile
    rm -rf processfile
    echo "------------------------- done $value"
    echo
    echo
    sleep 3 
  done

fi


exec /usr/local/bin/lotus $@
