# source this script to build versions of:
#   - open air simulator
# for plattforms:
#   - osx, linux, windows

VERSION=`git describe --tags --dirty`
DATE=`date +%Y%m%d`
LDFLAGS="-X main.version=${VERSION}_${DATE}"

REL_DIR=release
if [ ! -d ${REL_DIR} ]; then
	mkdir ${REL_DIR}
fi

for os in darwin linux windows; do
	GOOS=${os} GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${REL_DIR}/mqttGather.${VERSION}.${os} cmd/main.go
	GOOS=${os} GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${REL_DIR}/ca.${VERSION}.${os} cmd/ca/ca.go
done

