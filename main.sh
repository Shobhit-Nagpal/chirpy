#/bin/bash

echo "Starting server..."
go build -o out
mv out bin
./bin/out
