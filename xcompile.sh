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


go build -ldflags "${LDFLAGS}" -o ${REL_DIR}/mqttGather.${VERSION} cmd/main.go

# sqlite has native C bindings so won't work easily, TODO

#for os in linux windows; do
#	GOOS=${os} GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${REL_DIR}/mqttGather.${VERSION}.${os} cmd/main.go
#done
#
## PI
#GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "${LDFLAGS}" -o ${REL_DIR}/mqttGather.${VERSION}.linux.arm7 cmd/main.go
#
## MAC and MAC arm
#for arch in amd64 arm64; do
#	GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${REL_DIR}/mqttGather.${VERSION}.${os} cmd/main.go
#	GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${REL_DIR}/mqttGather.${VERSION}.${os}.arm cmd/main.go
#done
