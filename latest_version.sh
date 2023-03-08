#!/bin/bash

ver=$(git log -1  --abbrev-commit --date=format:'%Y%m%d%H%M%S' --pretty=format:"%H %ad %cd" | awk '{print "v0.0.0-"$2-80000"-"substr($1,0, 12)}')
repo=$(git remote get-url --push origin | awk '{gsub("git@", "", $0); gsub(":","/",$0); gsub(".git","",$0); print $0}')
echo "go get ${repo}@${ver}"