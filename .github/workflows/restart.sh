cd /var/www/html
echo "Changed directory to /var/www/html"

# Only kill if processes exist
pkill 'ingest'
echo "Killed ingest processes"

pkill 'serve'
echo "Killed serve processes"

nohup ./ingest > ingest.log 2>&1 &
echo "Started ingest process"

nohup ./serve > serve.log 2>&1 &
echo "Started serve process"