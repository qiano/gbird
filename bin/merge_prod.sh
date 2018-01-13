#!/bin/bash
# 合并到prod
# 分支合并规定为单向，顺序为：feature-xxx -> develop -> master -> test -> prod
# Author:   Daniel
# Date:     2017/12/11
# Version:  1.0

git checkout test && git pull
git checkout prod && git pull
git merge test --no-ff -m "合并test到prod：$(date '+%Y-%m-%d %T')"
git push origin
git checkout develop