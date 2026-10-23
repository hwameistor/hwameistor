#!/bin/sh
set -eu

export SSH_PRIVATE_KEY_PATH=/root/.ssh/id_rsa
exec juicesync --links --dirs --perms --force-update \
    "root@${sourceNode:?}:${sourceMountPoint:?}/" \
    "root@${targetNode:?}:${targetMountPoint:?}/"
