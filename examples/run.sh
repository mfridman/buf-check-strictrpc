#!/usr/bin/env bash

# Find the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Script directory: $SCRIPT_DIR"

# Try to find wasmtime in common locations
if [ -x "$HOME/.wasmtime/bin/wasmtime" ]; then
    echo "Found wasmtime in $HOME/.wasmtime/bin/wasmtime"
    WASMTIME_PATH="$HOME/.wasmtime/bin/wasmtime"
elif [ -x "/usr/local/bin/wasmtime" ]; then
    echo "Found wasmtime in /usr/local/bin/wasmtime"
    WASMTIME_PATH="/usr/local/bin/wasmtime"
else
    echo "Could not find wasmtime in common locations, using default path"
    WASMTIME_PATH="wasmtime"
fi

# Run the WASM binary using the detected wasmtime path
"$WASMTIME_PATH" "$SCRIPT_DIR/buf-check-strictrpc.wasm" "$@"
