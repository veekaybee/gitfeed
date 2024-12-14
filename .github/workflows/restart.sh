cd /var/www/html
chmod +x ingest serve
killall -0 ingest 2>/dev/null && pkill -f 'ingest'
killall -0 serve 2>/dev/null && pkill -f 'serve'
nohup ./ingest > ingest.log 2>&1 &
nohup ./serve > serve.log 2>&1 & 