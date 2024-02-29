# Fake

Fake is a Go type-safe mocking library!

Automatically create type-safe mocks for any public interface.

## Installation

### Go install

Make sure `$GO_PATH/bin` is in your `$PATH`

`go install github.com/sonalys/fake/entrypoints/fake@latest`

### Release download

Visit the [Release Page](https://github.com/sonalys/fake/releases) and download the correct binary for your architecture/os.

### Docker

You can safely run fake from a docker container using

`docker run --rm -u $(id -u):$(id -g) -v .:/code ghcr.io/sonalys/fake:latest /fake -input /code -output /code/mocks`

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
	setupWeirdFunc1 mockSetup.Mock[func(a any, b interface {
		A() int
	})]
	setupWeirdFunc2 mockSetup.Mock[func(in *<-chan time.Time, outs ...chan int) error]
	setupEmpty      mockSetup.Mock[func()]
	setupWeirdFunc3 mockSetup.Mock[func(a0 map[T]func(in ...*chan<- time.Time)) T]
}

func (s *StubInterface[T]) OnWeirdFunc1(funcs ...func(a any, b interface { A() int })) Config
func (s *StubInterface[T]) WeirdFunc1(a any, b interface { A() int })
...
```

So you can use it like this

```go

func Test_Stub(t *testing.T) {
  mock := mocks.NewStubInterface[int](t) // Setup call expectations
  mock.OnWeirdFunc1(func(a any, b interface { A() int }) {
    require.NotNil(t, a)
    ...
  })
  var Stub StubInterface[int] = mock
  Stub.WeirdFunc1(1, nil) // Will call the previous function.
}
```

You can pass more than one function or set repetition groups with

```go
mock.OnWeirdFunc1(func(a any, b interface { A() int }) {
  require.NotNil(t, a)
  ...
}).Repeat(2)
// or
mock.OnWeirdFunc1(func(a any, b interface { A() int }) {
  require.NotNil(t, a)
  ...
}).Maybe()
```