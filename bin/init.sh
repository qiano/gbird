#!/bin/bash
# 初始化分支结构
# Author:   Daniel
# Date:     2017/12/12
# Version:  1.0

git checkout master

branches=("develop" "test" "prod")
for i in ${!branches[@]}
do
    branch=${branches[$i]}
    git checkout -b $branch
    git push --set-upstream origin $branch
done

git checkout develop