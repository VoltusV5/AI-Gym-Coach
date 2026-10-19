#!/bin/sh
# Устанавливает закреплённый линтер в backend/bin и проверяет контрольную сумму релиза.
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=$(cat "$root/backend/.golangci-version")
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $(uname -m) in
  arm64|aarch64) arch=arm64 ;;
  x86_64) arch=amd64 ;;
  *) echo "Неподдерживаемая архитектура" >&2; exit 1 ;;
esac
archive="golangci-lint-${version}-${os}-${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
base="https://github.com/golangci/golangci-lint/releases/download/v${version}"
curl -fsSL --retry 3 "$base/$archive" -o "$tmp/$archive"
curl -fsSL --retry 3 "$base/golangci-lint-${version}-checksums.txt" -o "$tmp/checksums.txt"
cd "$tmp"
awk -v file="$archive" '$2 == file {print}' checksums.txt > selected.sha256
test -s selected.sha256
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum -c selected.sha256
else
  shasum -a 256 -c selected.sha256
fi
tar -xzf "$archive"
mkdir -p "$root/backend/bin"
cp "golangci-lint-${version}-${os}-${arch}/golangci-lint" "$root/backend/bin/golangci-lint"
