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

## test channel between command line thread and UDPServer thread
``` console
cd mp1/src/main
./main
```
some definitions of command  

NA means command line thread can directly get achieve those features

All commands which are needed to be processed in UDPserver can be executed in the next gossip period. (No hard real time requirement)

CHANGE_TO_ALL2ALL = 1  
CHANGE_TO_GOSSIP = 2  
LIST_MEMBERSHIPLIST = NA   
PRINT_SELF_ID = NA  
JOIN_GROUP = 3  
LEAVE_GROUP = 4  
FAIL = NA  

Please check test_output in mp1/src/main

## use ticker in UDPServer 

