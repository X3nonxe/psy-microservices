#!/bin/bash

PROTO_DIR=proto
GEN_DIR=gen

# Buat direktori gen jika belum ada
mkdir -p $GEN_DIR

# Generate kode Go
protoc --proto_path=$PROTO_DIR \
  --go_out=$GEN_DIR --go_opt=paths=source_relative \
  --go-grpc_out=$GEN_DIR --go-grpc_opt=paths=source_relative \
  $(find $PROTO_DIR -name '*.proto')