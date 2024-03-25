#!/usr/bin/env bash

GATE="$LOTUS_PATH"/date_initialized

if [ ! -z $DOCKER_LOTUS_IMPORT_SNAPSHOT ] ; then
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


cd /

exec /usr/local/bin/lotus $@
