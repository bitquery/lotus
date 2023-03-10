#!/bin/bash

if [ -z "$SNAPSHOT_URL" ]; then
   /usr/local/bin/lotus daemon      
else
   /usr/local/bin/lotus daemon  --import-snapshot $SNAPSHOT_URL
