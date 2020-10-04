## Structure
```
.
├── bootstrapMP1.sh (bootstrap script: change go path to mp1)
├── config.json (config file)
├── instruction (cs425 mp1 pdf)
├── README.md
├── report (cs425 group4 report)
└── src (source code)
```
## help
```
  -all2all
        start with all 2 all at the beginning
  -append
        append log rather than start a new log
  -clean int
        Cleanup Time; Remove the record from the membershiplist (default 3000)
  -config string
        Location of Config File (default "../../config.json")
  -fail int
        Fail Time (default 3300)
  -gossip int
        Gossip Period (default 300)
  -introducer
        start as an introducer
  -logLevel string
        log level: debug, info, warning, mute (default "debug")
  -loss int
        message loss rate 1-100 (default 1)
  -muteCli
        mute the command line interaction
  -port string
        Port used for Debug on one machine (default "1234")
  -simpleCli
        use simple cli
```


## How to use
```bash
cd mp1
source bootstrapMP1.sh
(1) cd src/main
    go run main.go -help
(2) runmp1
(3) debugmp1
```
 

