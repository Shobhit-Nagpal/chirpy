#!/bin/bash

# Build the Go project
go build -o out

# Get the name of the binary (assuming it's the same as the directory name)
binary_name="out"

# Check if 'bin' directory exists, create it if it doesn't
if [ ! -d "bin" ]; then
    mkdir bin
fi

# Move the binary to the 'bin' directory
mv "out" bin/

echo "Build complete. Binary moved to bin/out"

./bin/out
