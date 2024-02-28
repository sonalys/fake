package stub

import "time"

type StubInterface[T comparable] interface {
	Interf(a any, b interface {
		A() int
	})
	KillYourself(in *<-chan time.Time, outs ...chan int) error
	Empty()
	Weird(map[T]func(in ...*chan<- time.Time)) T
}
