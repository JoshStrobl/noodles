# Temporary Makefile for building noodles until we can build it with itself

ROOTDIR=$(shell pwd)

golang:
	cd $(ROOTDIR)/go && \
	GOPATH=$(ROOTDIR)/go go build -o $(ROOTDIR)/build/noodles src/noodles/*.go && \
	cd $(ROOTDIR)

.PHONY: golang
DEFAULT: golang