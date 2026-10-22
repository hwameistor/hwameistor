#!/bin/sh
# Run inside the candidate image: sh /tests/smoke.sh
set -eu
test_root=$(mktemp -d)
trap 'test_status=$?; if [ "$test_status" -ne 0 ]; then cat "$test_root"/*.log >&2 || :; fi; rm -rf "$test_root"; exit "$test_status"' EXIT
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

    # Repeating the production flags must also preserve existing dangling links.
    juicesync --links --dirs --perms --force-update --check-all \
        "$test_root/source/" "$test_root/$version/" > "$test_root/$version-repeat.log" 2>&1
    test "$(readlink "$test_root/$version/dangling")" = missing
    echo "PASS: $version repeated sync preserves dangling links"

    # Isolate regular-file checking from directory and symlink metadata behavior.
    mkdir "$test_root/$version-check-source" "$test_root/$version-check-target"
    source="$test_root/$version-check-source/file"
    target="$test_root/$version-check-target/file"
    printf 'migration content\n' > "$source"
    printf 'corrupted content\n' > "$target"
    touch -r "$source" "$target"
    juicesync --links --dirs --perms --check-new \
        "$test_root/$version-check-source/" "$test_root/$version-check-target/" > "$test_root/$version-check-new.log" 2>&1
    if cmp -s "$source" "$target"; then
        echo "check-new control unexpectedly repaired an unchanged-size/mtime file" >&2
        exit 1
    fi
    juicesync --links --dirs --perms --check-all \
        "$test_root/$version-check-source/" "$test_root/$version-check-target/" > "$test_root/$version-check-all.log" 2>&1
    cmp "$source" "$target"
    echo "PASS: $version check-all repairs different content with matching size and mtime"
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
