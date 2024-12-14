cd /var/www/html
echo "Changed directory to /var/www/html"

pkill -f 'ingest'
echo "Killed ingest processes"

pkill -f 'serve'
echo "Killed serve processes"

nohup ./ingest > ingest.log 2>&1 &
echo "Started ingest process"

nohup ./serve > serve.log 2>&1 &
echo "Started serve process"