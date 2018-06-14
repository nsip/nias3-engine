set -e
CWD=`pwd`

OUTPUT=$CWD/app/n3
cd app/n3
go get
go build
mkdir -p nss
cd $CWD
cd ../../nats-io/nats-streaming-server
GOOS=darwin
GOARCH=amd64
LDFLAGS="-s -w"
GNATS=nss/nats-streaming-server
#OUTPUT=$CWD/build/Mac/napval
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $OUTPUT/$GNATS
cd $CWD
cd n3
go get
go build
cd $CWD
cd n3crypto
go get
go build
cd $CWD
cd utils
go get
go build
cd $CWD
