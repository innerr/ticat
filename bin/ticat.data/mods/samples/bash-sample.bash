set -euo pipefail

echo "bash-sample in"

echo
echo "==> args:"
if [ ! -z "${1+x}" ]; then
	echo "- arg #1: ${1}"
fi
if [ ! -z "${2+x}" ]; then
	echo "- arg #2: ${2}"
fi
if [ ! -z "${3+x}" ]; then
	echo "- arg #3: ${3}"
fi
echo "---"

echo
input=`cat -`
if [ ! -z "${input}" ]; then
	echo "==> ticap env begin"
	echo "${input}"
	echo "---"
else
	echo "==> no content from stdin"
fi

echo
echo "==> read something from tty"
read msg </dev/tty
echo "got: ${msg}"
echo "---"

echo "write something to stderr 1" >&2
echo "write something to stderr 2" >&2

echo
echo "bash-sample out"
