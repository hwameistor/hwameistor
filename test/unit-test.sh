#! /usr/bin/env bash
# https://www.dazhuanlan.com/erlv/topics/1058820
set -x

# prepare output directory
COVERPKGS=( ./pkg/member/... ./pkg/controller/... ./pkg/utils/... )
tmp=$(mktemp -d)
merge="${tmp}/merge.out"
[ -f ${merge} ] && rm -f ${merge}

# go test for each package
for (( i=0; i<${#COVERPKGS[@]}; i++)) do
    ls $tmp
    cov_file="${tmp}/$i.cover"
    go test --race \
		--v \
		-covermode=atomic \
		-coverpkg=${COVERPKGS[i]} \
		-coverprofile=$cov_file \
		${COVERPKGS[i]}

	cat $cov_file | grep -v mode: | grep -v zz_generated  >> ${merge}
done

# merge all *.cover
header=$(head -n1 "${tmp}/0.cover")
echo "${header}" > coverage.out
cat ${merge} >> coverage.out
go tool cover -func=coverage.out
rm -rf coverage.out ${tmp}  ${merge}