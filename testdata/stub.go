package stub

import "time"

type StubInterface[T comparable] interface {
	AnotherInterface[T, time.Time]
	WeirdFunc1(a any, b interface {
		A() int
	})
	WeirdFunc2(in *<-chan time.Time, outs ...chan int) error
	Empty()
	WeirdFunc3(map[T]func(in ...*chan<- time.Time)) T
}

type AnotherInterface[J, A any] interface {
	DifferentGenericName(a J) A
}
