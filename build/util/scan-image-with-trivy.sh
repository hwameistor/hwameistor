#!/bin/bash
image_name=$1
arch_list=("amd64" "arm64")
severity="CRITICAL"
e_code=${2:-0}

# install trivy
{ which trivy 2>/dev/null; } || { echo "install trivy now..."; curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin latest; }

for arch in ${arch_list[@]}
do
    img_of_cur_arch=${image_name}-${arch}
    echo "# scan image for ${img_of_cur_arch}"

    # check image
    set +e
    docker inspect ${img_of_cur_arch} 2>&1 > /dev/null
    if [ $? -ne 0 ]; then
        echo "[Warning] No ${arch} image found on local , image name is ${img_of_cur_arch}. aborting  image manifest merge"
        exit 2
    fi
    set -e
    # scan image
    echo "# scaning image ${image_name}..."
    trivy image ${img_of_cur_arch} --severity ${severity} --exit-code ${e_code}

    # report
    echo "# scan successfully for image ${image_name}. No $severity level vulnerabilities found to be fixed."
done