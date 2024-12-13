cd /var/www/html && \
pkill -f 'go run cmd/ingest' || true && \
pkill -f 'go run cmd/serve' || true && \
nohup ./ingest > ingest.log 2>&1 & \
nohup ./serve > serve.log 2>&1 & 