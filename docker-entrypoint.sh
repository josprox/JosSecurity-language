#!/bin/bash
set -e

# Default to "joss server start" if no arguments provided
if [ $# -eq 0 ]; then
    set -- joss server start
fi

# Execute command
exec "$@"
