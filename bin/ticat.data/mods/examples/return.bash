set -euo pipefail

echo ">>> return-env example in"

echo
output="proto.ticat.env	display.width	120"
echo "==> modified session env by print values into stderr:"
echo "${output}"
echo "${output}" >&2

echo
echo "<<< return-env example out"
