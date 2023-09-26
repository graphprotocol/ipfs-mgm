# ipfs-mgm

Script to manage the IPFS files. It can be used to sync the CID's between two nodes

## Install

### Manually

```bash
go install cmd/cli/ipfs-mgm.go
```

## Usage
#### View help:

```bash
ipfs-mgm
```

or

```bash
ipfs-mgm --help
```

##### View subcommand help:

```bash
ipfs-mgm sync --help
```

#### Transfer all files from one IPFS node to another:

```bash
ipfs-mgm sync -s <SOURCE URL> -d <DESTINATION URL>
```

#### Transfer only specific files from one IPFS node to another:

```bash
ipfs-mgm sync -s <SOURCE URL> -d <DESTINATION URL> -f <FILE>
```

In this case, <FILE> has to be a file with one IPFS hash per line for each file that should be synced from the `-s` node to the `-d` node.

*Example*:

```text
QmZaHasmzsb1ReQHpCLUoXqqWchTgBrtRvbg7TqsUZXSuU
bafkreib47hfjivabsgaly3mupfqjvdywygfwnnoizlwg7wj2we3yh4t6fe
QmbyzKCFE6d22vnRVekN2Z5PT8Ha1g3TSku8UH5KBp2cTY
QmfNueFQg19hyBtCRUPJRpxVtdwtp8cgWpuRoQpRP3n9st
```

## Docker

The easiest way is to use the built docker image

```bash
docker run -it ghcr.io/graphprotocol/ipfs-mgm sync --help
```

## TODO:

- [ ] Implement async calls by creating a worker queue in batches
- [ ] Add directory support for sync operation
