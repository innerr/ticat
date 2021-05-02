set -euo pipefail

echo ">>> bash-sample in"

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

echo "==> use ticat to call other mods"
# Don't forget to quote "${input}"
ticat=`echo "${input}" | grep 'runtime.sys.paths.ticat' | tail -n 1 | awk '{print $3}'`
# Use stdin to pass the env to ticat
echo "${input}" | "${ticat}" v : sleep 1s

echo
echo "==> input anything and press enter:"
read msg </dev/tty
echo "got input from tty: '${msg}'"
echo "---"

echo
mod1="proto.ticat.env	samples.bash.input	${msg}"
mod2="proto.ticat.env	runtime.display.width	120"
mod3="(expected) wrong format context"
echo "==> modified session env by print values into stderr:"
echo "${mod1}"
echo "${mod2}"
echo "${mod3}"
echo "${mod1}" >&2
echo "${mod2}" >&2
echo "${mod3}" >&2
echo "---"

echo
echo "<<< bash-sample out"
