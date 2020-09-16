# Logly
Your friendly neighborhood distributed logging microservice from _Covfefe! Inc_.

## build logly on vm1
```
# Install Logly
cd $HOME
git clone https://github.com/beitong95/cs425.git
cd cs425/mp0
./bootstrap.sh
source ~/.bashrc
go build -o logly main
```

## run logly server one other machine
cd $HOME
git clone https://github.com/beitong95/cs425.git
cd cs425/mp0
./bootstrap.sh
.log_service_wrapper.sh

## Usage Examples
```
./logly --help
CONFIG=config.json ./logly -server &> serverlog.log &
CONFIG=config.json ./logly -client --expression="(A-Z)*" &> output.log
```

## Testing
```sh
go test ./...
go vet ./... # Check for program correctness
```
