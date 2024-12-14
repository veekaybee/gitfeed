#!/bin/bash

# Change to the application directory
cd /var/www/html

# Kill existing processes (suppress errors if processes don't exist)
pkill -f 'ingest' || true
pkill -f 'serve' || true

# Start services in background and redirect output to log files
nohup ./ingest > ingest.log 2>&1 &
nohup ./serve > serve.log 2>&1 &