# ipfs-mgm

Script to manage the IPFS files. It can be used to sync the CID's between two nodes

## Install

```bash
go install cmd/cli/ipfs-mgm.go
```

## Usage
Transfer all files from one IPFS node to another:

```bash
ipfs-mgm sync -s https://api.thegraph.com -d https://ipfs.thegraph.com
```

Transfer only specific files from one IPFS node to another:
```bash
ipfs-mgm sync -s <SOURCE URL> -d <DESTINATION URL> -f <FILE>
```

In this case, <FILE> has to be a file with one IPFS hash per line for each file that should be synced from the `-s` node to the `-d` node.

## Docker

- TO BE ADDED
