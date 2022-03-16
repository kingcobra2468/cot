#!/bin/bash

#pushd internal/gvoice
protoc --go-grpc_out=internal/text/gvoice --go_out=internal/text/gvoice --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative -I./internal/text/gvoice/proto/gvms gvoice.proto