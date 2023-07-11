# rpc-compare

A simple tool to compare the output of two RPC servers.

## Usage

`go run main.go --host1="http://127.0.0.1:8545" --host2="http://127.0.0.1:8546"`

That's it!  The files in the input folder will be read, sent to two rpc servers and the output will be compared.