# Server

Service API for sveltekit server data sourcing

## Setup

Pull code & install dependencies

```bash
go mod tidy
```


## Building

We use _Makefile_'s to chain together and manage build/deployment processes. Check the [`Makefile`](http://github.com/dd-web/opforu-server/blob/master/Makefile) for more details.

Build server binary

```bash
make build
```

## Running

After the binary is built you'll want to run it to listen for incoming requests from the sveltekit client server. We use the Makefile to expedite these processes and declare `build` a dependency of the `run` command, ensuring the binary is up to date every time it's ran:

```bash
make run
```

This will always build a fresh new binary before executing it.

