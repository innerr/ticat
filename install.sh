function check_cmd_exists()
{
	local cmd="${1}"
	if ! [ -x "$(command -v ${cmd})" ]; then
		echo "[:(] '"${cmd}"' not found, please install it first" >&2
		return 1
	fi
}

function download_bin()
{
	local repo="${1}"
	local bin_name="${2}"
	local dir="${3}"
	local token="Authorization: token ${4}"

	local os_type=`uname | awk '{print tolower($0)}'`
	echo "[:-] detected os type: ${os_type}" >&2

	local long_name="pre-builtin bin '${bin_name}' on '${repo}' for '${os_type}'"

	check_cmd_exists 'curl'

	local json=`curl --proto '=https' --tlsv1.2 -sSf -H "${token}" "https://api.${repo}/releases/latest"`

	local ver=`echo "${json}" | grep '"tag_name": ' | awk -F '"' '{print $(NF-1)}'`
	if [ -z "${ver}" ]; then
		echo "***" >&2
		echo "${json}" >&2
		echo "***" >&2
		echo "[:(] ${long_name} version not found or can't be downloaded, exiting" >&2
		return 1
	fi

	local info=`echo "${json}" | \
		grep '"assets": ' -A 99999 | \
		grep '"name": \|"browser_download_url": ' | \
		{ grep "${bin_name}_${os_type}" || test $? = 1; } | \
		awk -F '": "' '{print $2}' | \
		awk -F '"' '{print $1}'`

	local res_name=`echo "${info}" | { grep -v 'https://' || test $? = 0; }`
	if [ -z "${res_name}" ]; then
		echo "[:(] ${long_name} not found, exiting" >&2
		return 1
	fi

	local cnt=`echo "${res_name}" | wc -l | awk '{print $1}'`
	if [ "${cnt}" != '1' ]; then
		echo "***" >&2
		echo "${res_name}" | awk '{print "   - "$0}' >&2
		echo "***" >&2
		echo "[:(] error: more than one (${cnt}) resource of ${long_name}, exiting" >&2
		return 1
	fi

	local download_url=`echo "${info}" | tail -n 1`

	local hash_val=`echo "${res_name}" | awk -F '_' '{print $NF}' | awk -F '.' '{print $1}'`
	local hash_bin=`echo "${res_name}" | awk -F '_' '{print $(NF-1)}'`

	local bin_path="${dir}/${bin_name}"

	echo "[:)] located version ${ver}: ${long_name}"
	echo "   - ${hash_bin}: ${hash_val}"
	echo "   - url: ${download_url}"
	echo "   - download to: ${bin_path}"

	curl --proto '=https' --tlsv1.2 -sSf -kL -H "${token}" "${download_url}" > "${bin_path}"
	chmod +x "${bin_path}"

	echo "[:)] downloaded"
}

function download_and_install_ticat()
{
	local download_rate_limit_token="ghp_${1}"
	if [ -z "${2+x}" ]; then
		local mods_repo=''
	else
		local mods_repo="${2}"
	fi

	check_cmd_exists 'git'
	check_cmd_exists 'awk'

	local title='\033[1;94m'
	local green='\033[0;32m'
	local gray='\033[38;5;8m'
	local gray='\033[0;35m'
	local orange='\033[0;33m'
	local nc='\033[0m'

	local ticat_repo='github.com/repos/innerr/ticat'
	echo "${title}==> download ticat${nc}"
	download_bin "${ticat_repo}" 'ticat' '.' "${download_rate_limit_token}" 2>&1 | awk '{print "    * "$0}'

	echo
	if [ ! -z "${mods_repo}" ]; then
		echo "${title}==> fetch components from '${mods_repo}'${nc}"
		./ticat display.color.on : display.utf.off : display.width 90 : hub.add "${mods_repo}"  2>&1 | awk '{print "    * "$0}'
	else
		echo "${title}==> init ticat with basic component repos${nc}"
		./ticat display.color.on : display.utf.off : display.width 90 : hub.init  2>&1 | awk '{print "    * "$0}'
	fi

	echo
	echo "${title}==> add ticat to \$PATH${nc}"
	./ticat display.color.on : install.ticat  2>&1 | awk '{print "    * "$0}'

	echo
	echo "${green}==> Command ${orange}./ticat${green} is ready now, ${orange}ticat${green} is available after relogin${nc}"
	echo "    ${gray}ticat: workflow automating in unix-pipe style${nc}"
	echo "    ${gray}try:   $> ./ticat selftest"
}

set -eo pipefail
mods_repo="${1}"
set -u

download_and_install_ticat 'NYrOv0JuQ8iZ6cEnOTzdaTfh7ovx2Q2iwEQX' "${mods_repo}"
