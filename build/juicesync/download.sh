#!/bin/sh
set -eu

# Pinned official release archive checksums for both supported architectures.
arch=${1:?expected amd64 or arm64}
for version in 1.2.0 1.4.1; do
    case "$version/$arch" in
        1.2.0/amd64) checksum=28d6360c58b19627b07ea7c9f831c2162911f6eef7165289ff9bda465e256e0a ;;
        1.2.0/arm64) checksum=ef1781676feaffc0a007f09589e3c19f97d4299e90628f9830452ff2d4651d5f ;;
        1.4.1/amd64) checksum=01ee09a21a9351a465e09906f113845e1c6a19bea70f530e5e2b2125b2dd3b82 ;;
        1.4.1/arm64) checksum=1015ade83a7a93180a29f6c93ee5780a3eda52522331934ef2ef0cc0921995fd ;;
        *) echo "Unsupported architecture: $arch" >&2; exit 2 ;;
    esac
    archive="/tmp/juicefs-$version-linux-$arch.tar.gz"
    curl --fail --location --retry 3 --connect-timeout 30 --max-time 900 \
        "https://github.com/juicedata/juicefs/releases/download/v$version/juicefs-$version-linux-$arch.tar.gz" \
        --output "$archive"
    printf '%s  %s\n' "$checksum" "$archive" | sha256sum -c -
    mkdir -p "/opt/juicefs/$version"
    tar -xzf "$archive" -C "/opt/juicefs/$version" juicefs
    chmod +x "/opt/juicefs/$version/juicefs"
    rm "$archive"
done
