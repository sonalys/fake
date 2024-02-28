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


## Example

Running against our weird interface

```go
type StubInterface[T comparable] interface {
	WeirdFunc1(a any, b interface {
		A() int
	})
	WeirdFunc2(in *<-chan time.Time, outs ...chan int) error
	Empty()
	WeirdFunc3(map[T]func(in ...*chan<- time.Time)) T
}
```

Will generate the following mock:

```go
type StubInterface[T comparable] struct {
	setupWeirdFunc1 mock[func(a any, b interface {
		A() int
	})]
	setupWeirdFunc2 mock[func(in *<-chan time.Time, outs ...chan int) error]
	setupEmpty      mock[func()]
	setupWeirdFunc3 mock[func(a0 map[stub.T]func(in ...*chan<- time.Time)) stub.T]
}

func (s *StubInterface) OnWeirdFunc1(funcs ...func(a any, b interface { A() int })) Config
func (s *StubInterface) WeirdFunc1(a any, b interface { A() int })
...
```

So you can use it like this

```go

func Test_Stub(t *testing.T) {
  mock := mocks.NewStubInterface(t)
  mock.OnWeirdFunc1(func(a any, b interface { A() int }) {
    require.NotNil(t, a)
    ...
  })

  var Stub StubInterface[any] = mock

  Stub.WeirdFunc1(1, nil) // Will call the previous function.
}
```

You can pass more than one function or set repetition groups with

```go
mock.OnWeirdFunc1(func(a any, b interface { A() int }) {
  require.NotNil(t, a)
  ...
}).Repeat(2)
```