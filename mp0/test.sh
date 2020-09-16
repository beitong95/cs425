#! /bin/bash
a=`uname -n | sed -n 's/^.*g\(\S*\)-.*$/\1/p'`
echo $a
