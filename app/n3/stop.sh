# do not kill -9: that generates a SIGKILL which cannot be caught by the program, so it will not persist its stores
kill `cut -f 1 nats.pid`
rm nats.pid
echo "N3 services shut down"

