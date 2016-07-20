# Octarine

A transparent HTTP proxy visible only to wizards and cats.

## Info

Octarine functions both as a normal and transparent HTTP Proxy.

The Octarine "server" chooses a port to listen on randomly to minimize
start-up friction and because of this, has a "client" mode for the sole
purpose of querying for the port that the "server" is listening on.

The "client" and "server" pairing is made through the `<ID>` that is passed
in as a command line argument.

```
octarine [optional_flags] <ID>
```

For more information on available flags:
```
octarine -h
```

## Build

You can use `go build`, but if you want cross compilation then you'll need
to install `gox` (`go get github.com/mitchellh/gox`)

Requires `gox`:
```
make build
```

## Dev

This project adheres to the style dictated by standard Go tools (i.e. `gofmt`)

```
make lint
```
