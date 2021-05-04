set -euo pipefail

echo ">>> tty example in"

echo
echo "==> input anything and press enter:"
read msg </dev/tty
echo "got input from tty: '${msg}'"

echo
echo "<<< tty example out"
