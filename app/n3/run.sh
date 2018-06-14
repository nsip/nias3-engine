# This is the *nix NIAS batch file launcher. Add extra validators to the bottom of this list. 
# Change the directory as appropriate (go-nias)
# gnatsd MUST be the first program launched

if [ -f "nats.pid" ]
then
echo "There is a nats.pid file in place; run stop.sh"
exit
fi

#rem Run the NIAS services. Add to the BOTTOM of this list
# store each PID in pid list

#./nats-streaming-server -p 4223 -sc napval_nss.cfg & echo $! > nats.pid
nss/nats-streaming-server -p 4223 & echo $! > nats.pid

# give the nats server time to come up
sleep 5

./n3 -l 2000 -c SIF & echo $! >> nats.pid

echo "Launched NIAS3"

