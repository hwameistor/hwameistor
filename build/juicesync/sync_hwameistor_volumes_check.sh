#!/bin/sh
set -eu

export SSH_PRIVATE_KEY_PATH=/root/.ssh/id_rsa
exec juicesync --links --dirs --perms --force-update --check-all \
    "root@${sourceNode:?}:${sourceMountPoint:?}/" \
    "root@${targetNode:?}:${targetMountPoint:?}/"
