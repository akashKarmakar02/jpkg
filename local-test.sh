#!/bin/bash

# Ensure the target directory exists
mkdir -p /home/akash/.amber/bin

# Build the Go binaries
go build -o jpkg ./cmd
go build -o jpx runner/main.go

# Safely copy the built binaries
cp jpkg /home/akash/.amber/bin/jpkg.tmp
mv /home/akash/.amber/bin/jpkg.tmp /home/akash/.amber/bin/jpkg

cp jpx /home/akash/.amber/bin/jpx.tmp
mv /home/akash/.amber/bin/jpx.tmp /home/akash/.amber/bin/jpx

# Add to PATH if not already present
if ! grep -q 'export PATH=$PATH:/home/akash/.amber/bin' ~/.bashrc; then
    echo 'export PATH=$PATH:/home/akash/.amber/bin' >> ~/.bashrc
    echo "run source ~/.bashrc"
fi

# Update current session's PATH
export PATH=$PATH:/home/akash/.amber/bin
