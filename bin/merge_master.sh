#!/bin/bash
# 合并到master
# 分支合并规定为单向，顺序为：feature-xxx -> develop -> master -> test -> prod
# Author:   Daniel
# Date:     2017/12/11
# Version:  1.0

git checkout develop && git pull
git checkout master && git pull
git merge develop --no-ff -m "合并develop到master：$(date '+%Y-%m-%d %T')"
git push origin
git checkout develop