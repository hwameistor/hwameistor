SSH_PRIVATE_KEY_PATH=/root/.ssh/id_rsa juicesync --links --dirs --perms --force-update --check-new root@${sourceNode}:${sourceMountPoint}/ root@${targetNode}:${targetMountPoint}/
