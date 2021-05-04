set -euo pipefail

echo ">>> stderr example in"

echo
msg="(expected) error message"
echo "==> print content to stderr:"
echo "${msg}"
echo "${msg}" >&2

echo
echo "<<< stderr example out"
