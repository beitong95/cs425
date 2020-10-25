#!/bin/bash

echo "Clean go path setting in .bashrc"
sed -i "/\b\(GOPATH\)\b/d" ~/.bashrc
echo "Export GOPATH setting for MP1"
echo "export GOPATH=$HOME/cs425/mp2-revised:$HOME/cs425" >> ~/.bashrc

echo "Add runmp2 and debugmp2 alias"
sed -i "/\b\(runmp2\)\b/d" ~/.bashrc
echo "alias runmp2='cd ~/cs425/mp2-revised/src/main;go run main.go -logLevel info'" >> ~/.bashrc
echo "runmp2 = go run main.go -logLevel info"
sed -i "/\b\(debugmp2\)\b/d" ~/.bashrc
echo "alias debugmp2='cd ~/cs425/mp2-revised/src/main;go run main.go -simpleCli -logLevel debug'" >> ~/.bashrc
echo "debugmp2 = go run main.go -simpleCli -logLevel debug"

echo "Add VM env variable"
sed -i "/\b\(VMNUMBER\)\b/d" ~/.bashrc
VM=`uname -n | sed -n 's/^.*g04-\(\S*\)\.cs.*$/\1/p'`
echo "export VMNUMBER='vm'$VM" >> ~/.bashrc

echo "alias cleanRemoteFiles='rm -rf /home/beitong2/cs425/mp2-revised/files/*'" >> ~/.bashrc

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
