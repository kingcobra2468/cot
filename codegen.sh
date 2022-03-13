#!/bin/bash

#pushd internal/gvoice
protoc --go-grpc_out=internal/gvoice --go_out=internal/gvoice --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative -I./internal/gvoice/proto/gvms gvoice.proto