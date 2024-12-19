cd /var/www/html
echo "Changed directory to /var/www/html"

pkill 'serve'
echo "Killed serve processes"

nohup ./serve > serve.log 2>&1 &
echo "Started serve process"