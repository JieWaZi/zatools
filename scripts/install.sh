#!/usr/bin/env sh
set -eu

OWNER="${ZATOOLS_OWNER:-JieWaZi}"
REPO="${ZATOOLS_REPO:-zatools}"
BINARY_NAME="${ZATOOLS_BINARY:-zatools}"
VERSION="${VERSION:-}"
INSTALL_DIR="${ZATOOLS_INSTALL_DIR:-${INSTALL_DIR:-}}"

log() {
	printf '%s\n' "$*" >&2
}

fail() {
	log "error: $*"
	exit 1
}

download_file() {
	url="$1"
	output="$2"

	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$output"
		return
	fi
	if command -v wget >/dev/null 2>&1; then
		wget -qO "$output" "$url"
		return
	fi

	fail "curl or wget is required"
}

detect_os() {
	case "$(uname -s)" in
		Darwin) printf 'darwin' ;;
		Linux) printf 'linux' ;;
		*) fail "unsupported OS: $(uname -s)" ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64|amd64) printf 'amd64' ;;
		arm64|aarch64) printf 'arm64' ;;
		*) fail "unsupported architecture: $(uname -m)" ;;
	esac
}

sha256_file() {
	file="$1"

	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$file" | awk '{print $1}'
		return
	fi
	if command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$file" | awk '{print $1}'
		return
	fi
	if command -v openssl >/dev/null 2>&1; then
		openssl dgst -sha256 "$file" | awk '{print $NF}'
		return
	fi

	fail "missing sha256sum, shasum, or openssl"
}

resolve_install_dir() {
	if [ -n "$INSTALL_DIR" ]; then
		printf '%s' "$INSTALL_DIR"
		return
	fi

	printf '/usr/local/bin'
}

resolve_asset_from_checksums() {
	checksums_path="$1"
	preferred_asset="$2"
	os="$3"
	arch="$4"

	if [ -n "$preferred_asset" ]; then
		asset="$(awk -v asset="${preferred_asset}" '{name=$2; sub(/^\.\//, "", name); if (name == asset) {print name; exit}}' "${checksums_path}")"
		[ -n "$asset" ] || fail "unable to find checksum for ${preferred_asset}"
		printf '%s' "$asset"
		return
	fi

	status=0
	asset="$(awk -v prefix="${BINARY_NAME}_" -v suffix="_${os}_${arch}.tar.gz" '
		{
			name=$2
			sub(/^\.\//, "", name)
			if (index(name, prefix) == 1 && substr(name, length(name) - length(suffix) + 1) == suffix) {
				count++
				found=name
			}
		}
		END {
			if (count == 1) {
				print found
			} else if (count > 1) {
				exit 2
			} else {
				exit 1
			}
		}
	' "${checksums_path}")" || status=$?

	case "$status" in
		0) printf '%s' "$asset" ;;
		1) fail "unable to find a matching release asset in checksums.txt" ;;
		2) fail "multiple matching release assets found in checksums.txt" ;;
		*) fail "unable to read checksums.txt" ;;
	esac
}

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

os="$(detect_os)"
arch="$(detect_arch)"
install_dir="$(resolve_install_dir)"
if [ -n "$VERSION" ]; then
	base_url="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}"
	preferred_asset="${BINARY_NAME}_${VERSION}_${os}_${arch}.tar.gz"
else
	base_url="https://github.com/${OWNER}/${REPO}/releases/latest/download"
	preferred_asset=""
fi
checksums_path="${tmp_dir}/checksums.txt"
binary_path="${tmp_dir}/${BINARY_NAME}"
use_sudo=0

download_file "${base_url}/checksums.txt" "${checksums_path}"
asset="$(resolve_asset_from_checksums "${checksums_path}" "${preferred_asset}" "${os}" "${arch}")"
archive_path="${tmp_dir}/${asset}"

log "downloading ${asset}"
download_file "${base_url}/${asset}" "${archive_path}"

expected_hash="$(awk -v asset="./${asset}" '$2 == asset {print $1}' "${checksums_path}")"
if [ -z "$expected_hash" ]; then
	expected_hash="$(awk -v asset="${asset}" '$2 == asset {print $1}' "${checksums_path}")"
fi
[ -n "$expected_hash" ] || fail "unable to find checksum for ${asset}"

actual_hash="$(sha256_file "${archive_path}")"
[ "$expected_hash" = "$actual_hash" ] || fail "checksum verification failed for ${asset}"

tar -xzf "${archive_path}" -C "${tmp_dir}"
[ -f "${binary_path}" ] || fail "archive did not contain ${BINARY_NAME}"

if [ ! -w "${install_dir}" ]; then
	if [ "${install_dir}" = "/usr/local/bin" ] && command -v sudo >/dev/null 2>&1; then
		use_sudo=1
	else
		install_dir="${HOME}/.local/bin"
	fi
fi

if [ "$use_sudo" -eq 1 ]; then
	sudo mkdir -p "${install_dir}"
	sudo install -m 0755 "${binary_path}" "${install_dir}/${BINARY_NAME}"
else
	mkdir -p "${install_dir}"
	install -m 0755 "${binary_path}" "${install_dir}/${BINARY_NAME}"
fi

log "installed ${BINARY_NAME} to ${install_dir}/${BINARY_NAME}"
case ":$PATH:" in
	*:"${install_dir}":*)
		log "run: ${BINARY_NAME} --help"
		;;
	*)
		log "add ${install_dir} to PATH, then run: ${BINARY_NAME} --help"
		;;
esac
