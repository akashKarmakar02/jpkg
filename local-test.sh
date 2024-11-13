#!/bin/bash

# Ensure the target directory exists
mkdir -p /home/akash/.amber/bin

# Copy and rename files as per your original script
go build -o jpkg cmd/main.go
go build -o jpx runner/main.go

# Safely copy the built binaries
cp jpkg /home/akash/.amber/bin/jpkg.tmp
mv /home/akash/.amber/bin/jpkg.tmp /home/akash/.amber/bin/jpkg

cp jpx /home/akash/.amber/bin/jpx.tmp
mv /home/akash/.amber/bin/jpx.tmp /home/akash/.amber/bin/jpx
