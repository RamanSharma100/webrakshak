#!/usr/bin/env bash
set -euo pipefail
#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

echo "Building WebRakshak..."
# Build the main package (the repository root contains `main`), not all packages.
if ! go build -o webrakshak .; then
  echo "Build failed." >&2
  exit 1
fi

echo "Starting WebRakshak (logs -> webrakshak.log)..."
./webrakshak > webrakshak.log 2>&1 &
PID=$!
echo "WebRakshak started with PID $PID"
echo "To stop: kill $PID"
echo "To follow logs: tail -f webrakshak.log"

exit 0
