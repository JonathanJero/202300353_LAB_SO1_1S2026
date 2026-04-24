#!/bin/bash
docker run --rm -v $(pwd):/defs namely/protoc-all \
  -f wartweets.proto -l go -o .
