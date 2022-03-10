#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

#

#https://www.dazhuanlan.com/erlv/topics/1058820
set -x

#dir="("
#for i in `ls ./pkg/ | grep -v scheduler`;
#do
#  dir="$dir /pkg/$i/"
#  #echo "$dir"
#done
#
#dir="$dir )"

# ./pkg/...

PATH2TEST=( ./pkg/member/... ./pkg/controller/... ./pkg/common/... ./pkg/apis/... ./pkg/alerter/... ./pkg/exechelper/... ./pkg/utils/... )

tmpDir=$(mktemp -d)
mergeF="${tmpDir}/merge.out"
#rm -f ${mergeF}
for (( i=0; i<${#PATH2TEST[@]}; i++)) do
    ls $tmpDir
    echo ${#PATH2TEST[@]}
    cov_file="${tmpDir}/$i.cover"
    echo ${PATH2TEST[i]}
    go test --race --v  -covermode=atomic -coverpkg=${PATH2TEST[i]} -coverprofile=${cov_file}    ${PATH2TEST[i]}

    if [[ `cat $cov_file | grep -v mode: | grep -v zz_generated` = "" ]]
    then
      continue
    fi

    cat $cov_file | grep -v mode: | grep -v zz_generated  >> ${mergeF}

    #merge them
    header=$(head -n1 "${tmpDir}/$i.cover")
    echo "${header}" > coverage.out
    cat ${mergeF} >> coverage.out
done

#merge them
#header=$(head -n1 "${tmpDir}/0.cover")
#echo "${header}" > coverage.out
#cat ${mergeF} >> coverage.out
go tool cover -func=coverage.out
rm -rf coverage.out ${tmpDir}  ${mergeF}
