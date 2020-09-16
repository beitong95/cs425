#!/bin/bash

echo "export GOPATH=$HOME/cs425/mp0:$GOPATH" >> ~/.bashrc
source ~/.bashrc

a=`uname -n | sed -n 's/^.*g04-\(\S*\)\.cs.*$/\1/p'`
name='vm'$a'.test.log'
cp src/finder/machine.test.log $name

echo "Done."
