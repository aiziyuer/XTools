#!/bin/bash
execute() { echo "【command】 $@" ; eval "$@" ; }
info() { echo "【info】 $@" ; }

# 项目根目录
export WORK_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && cd ../ && pwd )"

execute "export PLATFORMS='windows linux darwin'"
execute "export ARCHS=amd64" 

info '////////////// compile start ///////////////'

for APP in $(ls cmd); do
    for PLATFORM in ${PLATFORMS}; do
        for ARCH in ${ARCHS}; do
        
            execute "export APP=${APP}"
            execute "export GOOS=${PLATFORM} GOARCH=${ARCH}"
            execute "export RELEASE='${APP}-${GOOS}-${GOARCH}'"
            execute "export BUILD_DIR='${WORK_DIR}/output/${RELEASE}'"

            # 在项目根目录构建
            execute "mkdir -p ${BUILD_DIR}"
            execute "cp -rf $WORK_DIR/configs ${BUILD_DIR}/"
            execute "export GO111MODULE=on"
            execute "export GOPROXY=https://nexus.moyi-lc.com:5003/repository/go-group/"
            execute "cd $WORK_DIR && go clean -modcache && go build -v -o ${BUILD_DIR}/${APP} app/cmd/${APP}"
            execute "cd ${BUILD_DIR} && tar czvf ../${RELEASE}.tar.gz *"

        done
    done
done
