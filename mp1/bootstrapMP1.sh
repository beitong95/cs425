#!/bin/bash

echo "Clean go path setting in .bashrc"
sed -i "/\b\(GOPATH\)\b/d" ~/.bashrc
echo "Export GOPATH setting for MP1"
echo "export GOPATH=$HOME/cs425/mp1:$HOME/cs425" >> ~/.bashrc
source ~/.bashrc
echo "Current GOPATH: "
env | grep GOPATH

#a=`uname -n | sed -n 's/^.*g04-\(\S*\)\.cs.*$/\1/p'`
#name='vm'$a'.test.log'
#cp src/finder/machine.test.log $name

echo "Done."
