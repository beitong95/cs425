#! /bin/bash
cd /home/beitong2
rm -rf cs425/
sleep 1
git clone https://github.com/beitong95/cs425.git&
sleep 5
cd cs425
git submodule update --init --recursive
sleep 5
