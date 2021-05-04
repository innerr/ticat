set -euo pipefail

echo ">>> call-mod example in"

input=`cat -`
echo "==> use ticat to call other mods"
# Don't forget to quote "${input}"
ticat=`echo "${input}" | grep 'sys.paths.ticat' | tail -n 1 | awk '{print $3}'`
# Use stdin to pass the env to ticat
echo "${input}" | "${ticat}" v : sleep 1s

echo
echo "<<< call-mod example out"
