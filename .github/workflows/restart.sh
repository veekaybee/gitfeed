#!/bin/bash

cd /var/www/html


pkill -f 'ingest' || true
pkill -f 'serve' || true


nohup ./ingest > ingest.log 2>&1 &
nohup ./serve > serve.log 2>&1 &