#!/bin/bash
echo "Building JosSecurity Distribution..."
go run cmd/dist/main.go
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi
echo "Build complete."