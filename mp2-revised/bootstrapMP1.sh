#!/bin/bash

echo "Clean go path setting in .bashrc"
sed -i "/\b\(GOPATH\)\b/d" ~/.bashrc
echo "Export GOPATH setting for MP1"
echo "export GOPATH=$HOME/cs425/mp1:$HOME/cs425" >> ~/.bashrc

echo "Add runmp1 and debugmp1 alias"
sed -i "/\b\(runmp1\)\b/d" ~/.bashrc
echo "alias runmp1='cd ~/cs425/mp1/src/main;go run main.go -logLevel info'" >> ~/.bashrc
echo "runmp1 = go run main.go -logLevel info"
sed -i "/\b\(debugmp1\)\b/d" ~/.bashrc
echo "alias debugmp1='cd ~/cs425/mp1/src/main;go run main.go -simpleCli -logLevel debug'" >> ~/.bashrc
echo "debugmp1 = go run main.go -simpleCli -logLevel debug"

echo "Add VM env variable"
sed -i "/\b\(VMNUMBER\)\b/d" ~/.bashrc
VM=`uname -n | sed -n 's/^.*g04-\(\S*\)\.cs.*$/\1/p'`
echo "export VMNUMBER='vm'$VM" >> ~/.bashrc

source ~/.bashrc

echo "Current GOPATH: "
env | grep GOPATH
echo "Current VMNUMBER: "
env | grep VMNUMBER

echo "Get submodules"
cd $HOME/cs425
git submodule update --init --recursive
#name='vm'$a'.test.log'
#cp src/finder/machine.test.log $name

echo "Done."
