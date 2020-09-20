## test print membership list in the table format & get local ip address
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
you can create your own test case in printTable_Test.go and getIP_test.go

sample output can be found in mp1/src/helper/test_output
```
------------------------------------------
|ID       |HeartBeat   |LocalTime        |
|---------|------------|-----------------|
|1        |1           |1                |
|cs425    |2           |3                |
|mp1      |3           |1111111111111    |
------------------------------------------
```
