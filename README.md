# OPforU API Server

Back end code for OPforU project. Service API for client server data sourcing.

## Setup

Once you have pulled down the code you need to install the dependencies from the root directory:

```bash
go mod tidy
```

This will fetch any dependencies necessary for the project. Depending on which IDE or LSP you use warnings may appear without them.

## Building

We use _Makefile_'s to chain together and manage build/deployment processes. Check the [`Makefile`](http://github.com/dd-web/opforu-server/blob/master/Makefile) for more details.

To build the binary:

```bash
make build
```

## Running

After the binary is built you'll want to run it to listen for incoming requests. We use the Makefile to expedite these processes and declare `build` a dependent of the `run` command, ensuring the binary is up to date every time it's ran:

```bash
make run
```

This will always build a fresh new binary before executing it.

