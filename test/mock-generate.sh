#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

#

#https://www.dazhuanlan.com/erlv/topics/1058820
set -x

FIND_STR="go:generate mockgen"

function makeGenerate()
{
  cd $1 && go generate
  cd -
}

function cleanUpMocks()
{
  rm -rf $1/mocks
}

function listFiles()
{
        #1st param, the dir name
        #2nd param, the aligning space
        alignSpace=$2
        if [ ! -n "$alignSpace" ]; then
          echo "alignSpace IS NULL"
          alignSpace=""
        fi

        for file in `ls $1`;
        do
                if [ -d "$1/$file" ]; then
                    echo "$alignSpace$file"
                    listFiles "$1/$file" " $alignSpace"
                else
                    # 判断匹配函数，匹配函数不为0，则包含给定字符
                    if [ `grep -c "$FIND_STR" $1/$file` -ne '0' ];then
                        echo "The File Has 'go:generate mockgen'!"
                        makeGenerate $1
                    fi
                    echo "$1/$file"
                fi
        done
}

cleanUpMocks $1
listFiles $1 ""


