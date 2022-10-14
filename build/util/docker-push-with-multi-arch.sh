#!/bin/bash
set -e
while getopts "i:s:" opt;
do
	case $opt in
		i)
		image_name=$OPTARG
		;;

		s)
		scan_image_before_push=$OPTARG
		;;

		?)
		echo "unknow params"
		exit 1;;
	esac
done

PAOGRAM=$(cd `dirname $0`; pwd)
arch_list=("amd64" "arm64")
scan_image_exit_code=1

# scan image before push if set true
if [ "${scan_image_before_push}" == "true" ];then
  { bash -x ${PAOGRAM}/scan-image-with-trivy.sh ${image_name} ${scan_image_exit_code}; } || { echo "[Error] scan image ${image_name} fail"; exit 2; }
fi

for arch in ${arch_list[@]}
do
    echo "# build manifest for ${arch}"
    img_of_cur_arch=${image_name}-${arch}

    # check image
    set +e
    docker inspect ${img_of_cur_arch} 2>&1 > /dev/null
    if [ $? -ne 0 ]; then
        echo "[Warning] No ${arch} image found on local , image name is ${img_of_cur_arch}. aborting  image manifest merge"
        exit 2
    fi 
    set -e
    
    # push image
    echo "[1/3] push image ${img_of_cur_arch}  to hub $hub.."
    docker push ${img_of_cur_arch}; 
    
    # creste manifest
    echo "[2/3] add local newly build ${arch} image into manifest .."
    docker manifest create --insecure --amend ${image_name}  ${img_of_cur_arch} ;

    # annotate manifest
    echo "[3/3] add annotation image into manifest .."
    docker manifest annotate ${image_name} ${img_of_cur_arch} --os linux --arch ${arch}
done

# push manifest
echo "[Info] pushing manifests .."
docker manifest push --insecure --purge   ${image_name} ;
echo "done"
