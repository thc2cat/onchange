include ../make/Makefile-for-go.mk 

REMOTE_DESTINATION= "root@logs-siege.csi.uvsq.fr.:/local/src/bin/"
NAME= $(notdir $(shell pwd))
TAG=$(shell git tag)

build:
	@go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ${NAME}-${TAG}
	@notify-send 'Build Complete' 'Your version has been updated successfully!' -u normal -t 7500 -i checkbox-checked-symbolic

