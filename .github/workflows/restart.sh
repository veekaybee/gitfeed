cd /var/www/html
echo "Changed directory to /var/www/html"

# Only kill if processes exist
pkill -f 'ingest'
echo "Killed ingest processes"

pkill -f 'serve'
echo "Killed serve processes"

# Wait a moment for processes to clean up
sleep 2
echo "Waited for cleanup"

nohup ./ingest > ingest.log 2>&1 &
echo "Started ingest process"

nohup ./serve > serve.log 2>&1 &
echo "Started serve process"