#!/bin/sh
# survive85 installer.
#
#   curl -fsSL https://raw.githubusercontent.com/funstorm15/survive85/main/scripts/install.sh | sh
#
# Environment:
#   VERSION=v1.2.3      install a specific release (default: latest)
#   INSTALL_DIR=/path   where to put the binary (default: ~/.local/bin)
#   USE_SUDO=1          allow sudo when INSTALL_DIR is not writable
set -eu

REPO="funstorm15/survive85"
BINARY="survive85"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

say() { printf '\033[36m==>\033[0m %s\n' "$*"; }
die() { printf '\033[31merror:\033[0m %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not installed"; }

need uname
need tar
if command -v curl >/dev/null 2>&1; then
	fetch() { curl -fsSL "$1" -o "$2"; }
	resolve_latest() { curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest"; }
elif command -v wget >/dev/null 2>&1; then
	fetch() { wget -qO "$2" "$1"; }
	resolve_latest() { wget -q --max-redirect=5 -S --spider "https://github.com/$REPO/releases/latest" 2>&1 | sed -n 's/^ *[Ll]ocation: //p' | tail -n1 | tr -d '\r'; }
else
	die "curl or wget is required"
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
	linux | darwin) ;;
	*) die "unsupported OS: $os (Linux and macOS are supported)" ;;
esac

arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	armv7l | armv7) arch=armv7 ;;
	*) die "unsupported architecture: $arch" ;;
esac
[ "$os" = darwin ] && [ "$arch" = armv7 ] && die "unsupported platform: darwin/armv7"

if [ "$VERSION" = latest ]; then
	url=$(resolve_latest) || die "cannot reach GitHub to find the latest release"
	VERSION=${url##*/}
	[ -n "$VERSION" ] && [ "$VERSION" != latest ] || die "cannot determine the latest release"
fi
num=${VERSION#v}
archive="${BINARY}_${num}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$VERSION"

tmp=$(mktemp -d 2>/dev/null || mktemp -d -t survive85)
trap 'rm -rf "$tmp"' EXIT INT TERM

say "Downloading $BINARY $VERSION for $os/$arch"
fetch "$base/$archive" "$tmp/$archive" || die "download failed: $base/$archive"
fetch "$base/checksums.txt" "$tmp/checksums.txt" || die "download failed: checksums.txt"

say "Verifying checksum"
expected=$(grep " $archive\$" "$tmp/checksums.txt" | cut -d' ' -f1)
[ -n "$expected" ] || die "$archive is not listed in checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
	actual=$(sha256sum "$tmp/$archive" | cut -d' ' -f1)
elif command -v shasum >/dev/null 2>&1; then
	actual=$(shasum -a 256 "$tmp/$archive" | cut -d' ' -f1)
else
	die "sha256sum or shasum is required to verify the download"
fi
[ "$expected" = "$actual" ] || die "checksum mismatch for $archive"

tar -xzf "$tmp/$archive" -C "$tmp" "$BINARY" || die "cannot unpack $archive"

say "Installing to $INSTALL_DIR/$BINARY"
if mkdir -p "$INSTALL_DIR" 2>/dev/null && [ -w "$INSTALL_DIR" ]; then
	install -m 755 "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
elif [ "${USE_SUDO:-0}" = 1 ] && command -v sudo >/dev/null 2>&1; then
	sudo mkdir -p "$INSTALL_DIR"
	sudo install -m 755 "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
else
	die "$INSTALL_DIR is not writable (set INSTALL_DIR, or USE_SUDO=1)"
fi

case ":$PATH:" in
	*":$INSTALL_DIR:"*) ;;
	*) say "Note: $INSTALL_DIR is not in your PATH. Add it, e.g.: export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac
say "Done. Run '$BINARY' to start living. ('$BINARY -lang ru' for Russian)"
