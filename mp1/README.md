## test print membership list in table format
For first time usage, you need to add the new go path
```console
cd mp1/src/
bash ./bootstrap.sh
source ~/.bashrc
```

```console
cd mp1/src/helper/
go test -v
```
you can create your own test case in printTable_Test.go

sample output

ID            HeartBeat     LocalTime
------------- ------------- -------------
1             1             1
cs425         1.123         1.123
mp1           1.123         1.123
