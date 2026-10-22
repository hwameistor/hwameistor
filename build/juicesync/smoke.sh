#!/bin/sh
# Run inside the candidate image: sh /tests/smoke.sh
set -eu
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT
mkdir -p "$test_root/source/directory" "$test_root/source/empty"
printf 'migration content\n' > "$test_root/source/directory/file with spaces"
ln -s directory "$test_root/source/link"
ln -s missing "$test_root/source/dangling"

for version in default 1.2.0 1.4.1; do
    if [ "$version" = default ]; then
        unset JUICEFS_VERSION
        expected=1.2.0
    else
        export JUICEFS_VERSION=$version
        expected=$version
    fi
    log="$test_root/$version.log"
    juicesync --links --dirs --perms --force-update --check-all \
        "$test_root/source/" "$test_root/$version/" > "$log" 2>&1
    grep -q "JuiceFS $expected, check=--check-all" "$log"
    cmp "$test_root/source/directory/file with spaces" "$test_root/$version/directory/file with spaces"
    test "$(readlink "$test_root/$version/link")" = directory
    test "$(readlink "$test_root/$version/dangling")" = missing
    test -d "$test_root/$version/empty"
    echo "PASS: $version content, links, empty directory, check-all"
done

if JUICEFS_VERSION=invalid juicesync --help > "$test_root/invalid.log" 2>&1; then
    echo "Invalid version unexpectedly succeeded" >&2
    exit 1
else
    test "$?" = 2
fi
echo "PASS: invalid version fails without fallback"

# Unsupported objects must retain the documented difference between versions.
mkdir "$test_root/special"
mkfifo "$test_root/special/pipe"
if JUICEFS_VERSION=1.2.0 juicesync --links --dirs --check-all \
    "$test_root/special/" "$test_root/special-1.2.0/" > "$test_root/special-1.2.0.log" 2>&1; then
    echo "JuiceFS 1.2.0 unexpectedly accepted a FIFO" >&2
    exit 1
fi
JUICEFS_VERSION=1.4.1 juicesync --links --dirs --check-all \
    "$test_root/special/" "$test_root/special-1.4.1/" > "$test_root/special-1.4.1.log" 2>&1
test ! -e "$test_root/special-1.4.1/pipe"
grep -q 'skips special files' "$test_root/special-1.4.1.log"
echo "PASS: 1.2.0 fails on FIFO; 1.4.1 explicitly warns and skips it"
