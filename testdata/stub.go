package stub

import (
	"testing"
	"time"

	"github.com/sonalys/fake/testdata/anotherpkg"
	stub "github.com/stretchr/testify/require"
)

type LocalType struct{}

type StubInterface[T comparable] interface {
	anotherpkg.ExternalInterface
	AnotherInterface[T, time.Time]
	WeirdFunc1(a any, b interface {
		A() int
	})
	WeirdFunc2(in *<-chan time.Time, outs ...chan int) error
	Empty()
	WeirdFunc3(map[T]func(in ...*chan<- time.Time)) T
	External(testing.T, stub.Assertions)
	LocalType(LocalType)
}

type AnotherInterface[J, A any] interface {
	DifferentGenericName(a J) A
}
