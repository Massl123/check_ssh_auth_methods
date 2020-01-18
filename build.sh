#!/bin/bash

releaseDirs=("dist/check_ssh_auth_methods-linux-amd64" "dist/check_ssh_auth_methods-linux-386" "dist/check_ssh_auth_methods-linux-arm" "dist/check_ssh_auth_methods-linux-arm64")

if [[ "$(pwd)" != *"/go/src/github.com/massl123/check_ssh_auth_methods" ]]
then
    echo "Switch to correct path (~/go/src/github.com/massl123/check_ssh_auth_methods)"
    exit 1
fi

version="$1"
if [[ ! $1 ]]
then
    echo "Provide release version!"
    echo "./build.sh v0.1.0"
    exit 1
fi

git checkout tags/${version}
ec=$?
if [[ ${ec} != 0 ]]
then
    echo "Checkout did not work!"
    exit 1
fi


echo "Removing old builds"
rm -rf dist/*

echo "Creating folders and copying elements"
for dir in ${releaseDirs[*]}
do
    mkdir -p ${dir}
    cp LICENSE ${dir}
    cp README.md ${dir}
done

echo "Building..."
GOOS=linux GOARCH=amd64 go build -o dist/check_ssh_auth_methods-linux-amd64
GOOS=linux GOARCH=386 go build -o dist/check_ssh_auth_methods-linux-386
GOOS=linux GOARCH=arm go build -o dist/check_ssh_auth_methods-linux-arm
GOOS=linux GOARCH=arm64 go build -o dist/check_ssh_auth_methods-linux-arm64


echo "Creating tarballs"
mkdir dist/final
tar czpf dist/final/check_ssh_auth_methods-${version}-linux-amd64.tgz -C dist/check_ssh_auth_methods-linux-amd64 .
tar czpf dist/final/check_ssh_auth_methods-${version}-linux-386.tgz -C dist/check_ssh_auth_methods-linux-386 .
tar czpf dist/final/check_ssh_auth_methods-${version}-linux-arm.tgz -C dist/check_ssh_auth_methods-linux-arm .
tar czpf dist/final/check_ssh_auth_methods-${version}-linux-arm64.tgz -C dist/check_ssh_auth_methods-linux-arm64 .

echo "Switching back to master branch"
git checkout master
