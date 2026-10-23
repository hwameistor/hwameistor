# Migration image

The image bundles the official Linux JuiceFS 1.2.0 and 1.4.1 binaries. `/usr/bin/juicesync` selects `/opt/juicefs/${JUICEFS_VERSION}/juicefs` and invokes its `sync` subcommand. The default is `1.2.0`. There is no fixed version list: the selected binary must exist and be executable. The version identifies a single directory under `/opt/juicefs`; path components such as `/`, `.` and `..` are rejected. Missing or non-executable binaries fail with exit code 2, and there is no automatic download or fallback.

HwameiStor migration keeps `--links --dirs --perms --force-update`. Local Storage now defaults `--migrate-check` to true, selecting `sync_hwameistor_volumes_check.sh`, which adds `--check-all`. This verifies copied files as well as existing files that would otherwise be skipped based on size and modification time. Explicit false selects the non-checking script. Direct invocations of the `juicesync` wrapper must supply their own sync flags.

For a direct HwameiStor Helm installation:

```yaml
localStorage:
  migrate:
    check: true
    juicesync:
      imageRepository: hwameistor/hwameistor-juicesync
      tag: v1.2.0
      version: "1.2.0"
```

The tag above is the release target for this change; building the code does not publish it. Deploy with the corresponding updated Local Storage image. The controller passes `JUICEFS_VERSION` explicitly into each new sync Job. `MIGRAGE_JUICESYNC_IMAGE` retains its existing spelling and controls the container image, independently of the selected binary. Helm, the controllers and the wrapper do not enumerate allowed versions. For example, selecting `1.5.0` requires an executable at `/opt/juicefs/1.5.0/juicefs`; adding it to a custom image requires no controller or wrapper changes. The image built here still includes only `1.2.0` and `1.4.1`.

For Operator-managed installations, the companion Operator change uses the same Helm values to set `JUICEFS_VERSION` and `MIGRATE_CHECK` on the Operator Deployment. Operator forwards the selected version to Local Storage and reconciles its `--migrate-check` argument. Operator environment variables provide defaults for its managed Clusters. The existing `localStorage.member.extraEnv` Helm value (stored in the `hwameistor.io/local-storage-member-extra-env` Cluster annotation) can override `JUICEFS_VERSION` for an individual Cluster. User environment entries retain their declaration order, including `$(VAR_NAME)` references and `valueFrom`; Kubernetes resolves them when starting the Local Storage container. The last version entry wins; removing all version entries restores the Operator default. `MIGRATE_CHECK` is configured on the Operator Deployment. No Cluster API or CRD fields are added. For deployments managed without Helm:

```sh
kubectl -n <operator-namespace> set env deployment/hwameistor-operator JUICEFS_VERSION=1.4.1
# Optional: explicitly disable copied-content verification.
kubectl -n <operator-namespace> set env deployment/hwameistor-operator MIGRATE_CHECK=false
```

The existing Cluster `spec.localStorage.member.juicesyncImage` field continues to select the image. Configure the Operator Deployment instead of editing its managed Local Storage DaemonSet. Upgrade Operator and Local Storage together with the migration image; no CRD upgrade is required. Old controllers do not forward the version variable, and custom old migration images do not implement the selector. Wait for the Local Storage rollout to finish before creating a new migration Job. Existing Jobs, including their retry Pods, retain their original Pod template; these settings apply to newly created Jobs.

JuiceFS 1.2.0 fails on the tested FIFO, socket and device objects. Version 1.4.1 warns and skips these objects, including character-device OverlayFS whiteouts; `--check-new` and `--check-all` cannot detect objects filtered before copying. Selecting 1.4.1 is an explicit change to that behavior. Neither version preserves hardlink relationships or xattrs/ACLs through this SFTP sync path. Content verification is not a complete rootfs integrity check and does not replace stopping writes or using a consistent source snapshot. It adds source/target read and network I/O.

Run shell tests without a cluster:

```sh
python3 build/juicesync/test_scripts.py
go test ./pkg/local-storage/utils/datacopy
```

Build and test a local image without publishing:

```sh
docker build -f build/juicesync/Dockerfile.amd64 -t hwameistor-juicesync:test .
docker run --rm -v "$PWD/build/juicesync/smoke.sh:/tests/smoke.sh:ro" \
  hwameistor-juicesync:test sh /tests/smoke.sh
```

Use `Dockerfile.arm64` and a suitable builder for ARM64. `JUICESYNC_TAG` in the Makefile only controls the image release tag; every build includes both binaries regardless of that tag. The existing `make juicesync` release target builds both architectures and publishes the multi-architecture image. Tests are run manually using the commands above and are not part of the image build or release flow. Archive SHA-256 values are pinned in `download.sh`; no binaries are downloaded at container startup.
