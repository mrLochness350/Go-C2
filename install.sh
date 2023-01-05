#!/bin/bash

echo "Creating bin/ directory..."
if [[ ! -d  "bin" ]]
then 
    mkdir bin
fi

echo "Installing dependencies..."
go get -v github.com/akamensky/argparse
go get -v github.com/armon/go-socks5

echo "Running make command..."
make Makefile build



