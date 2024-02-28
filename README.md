# Fake

Fake is a Go type-safe mocking library!

Automatically create type-safe mocks for any public interface.

## Installation

### Go install

Make sure `$GO_PATH/bin` is in your `$PATH`

`go install github.com/sonalys/fake/entrypoints/cli@latest`

### Release download

Visit the [Release Page](https://github.com/sonalys/fake/releases) and download the correct binary for your architecture/os.

### Docker

You can safely run fake from a docker container using

`docker run --rm ghcr.io/sonalys/fake:latest fake -input PATH -output mocks -ignore tests`

## Usage

```

Usage:

  fake -input PATH -output mocks -ignore tests

The flags are:

  -input    STRING    Folder to scan for interfaces, can be invoked multiple times
  -output   STRING    Output folder, it will follow a tree structure repeating the package path
  -ignore   STRING    Folder to ignore, can be invoked multiple times
```