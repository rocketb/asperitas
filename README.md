# Asperitas: reddit clone backend

## Description
Asperitas is an implementation of the [Asperitas](https://github.com/d11z/asperitas) project, which draws inspiration from Reddit and is built entirely in Go.
It serves as the backend component for the Asperitas project.
Please note that this project is not fully compatible with the original Asperitas frontend.

## Building from source
To build this project from source, you will need Go.
You can either use the Go version specified in the provided Dockerfile or install a compatible version on your own.
```
$ go get github.com/rockeb/asperitas
$ cd $GOPATH/src/github.com/rocketb/asperitas # GOPATH is $HOME/go by default.

$ go build ./cmd/asperitas/api
...

```

To build metrics exporter use the following command:
```
go build ./cmd/asperitas/metrics
```

To build admin tools cli tool use the following command:
```
go build ./cmd/tools/asperitas-admin
```

Please read the [Makefile](Makefile) file to learn how to install all the tooling and more.
