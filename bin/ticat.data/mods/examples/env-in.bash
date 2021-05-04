set -euo pipefail

echo ">>> env example in"

echo
input=`cat -`
if [ ! -z "${input}" ]; then
	echo "==> ticat env from stdin begin"
	echo "${input}"
else
	echo "==> no content from stdin"
fi

echo
echo "<<< env example out"
