# Fake

Fake is a Go type-safe [mocking](https://en.wikipedia.org/wiki/Mock_object) generator. It automatically creates type-safe mocks for any public interface.

## Features

- Type-safe mock generation
- Support for generics
- Mock cache for ultra-fast mock regeneration
- Function call configuration, with Repeatability and Optional calls
- Automatic call assertion

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

A very simple example would be:

```go
type UserDB interface {
	Login(userID string) error
}
```

Will generate the following mock:

```go
type UserDB interface {
	Login(userID string) error
}
type StubInterface struct {
	setupLogin      mockSetup.Mock[func(userID string) error]
}

func (s *StubInterface[T]) OnLogin(funcs ...func(userID string) error) Config
func (s *StubInterface[T]) Login(userID string) error
...
```

---

```go

func Test_Stub(t *testing.T) {
  mock := mocks.NewUserDB(t) // Setup call expectations
  config := mock.OnLogin(func(userID string) error {
    require.NotEmpty(t, userID)
    return nil
  })
  // Here you can configure it with:
  config.Repeat(1) // Repeat 1 is default behavior.
  // or with .Maybe(), which won't fail tests it not called.
  config.Maybe()

  var userDB UserDB = mock
  userDB.Login("userID") // Will call the previous function.
}
```

---

## Contributors

Any issues or improvements discussions are welcome! Feel free to contribute.
Thank you for your collaboration!

<a href="https://github.com/sonalys/fake/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=sonalys/fake" />
</a>

---
