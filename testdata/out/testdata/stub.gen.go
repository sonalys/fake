// Code generated by mockgen. DO NOT EDIT.

package mocks

import (
	"fmt"
	_ "github.com/sonalys/fake/boilerplate"
	"github.com/sonalys/fake/testdata"
	"testing"
	"time"
)

type StubInterface[T comparable] struct {
	setupWeirdFunc1 mock[func(a any, b interface {
		A() int
	})]
	setupWeirdFunc2 mock[func(in *<-chan time.Time, outs ...chan int) error]
	setupEmpty      mock[func()]
	setupWeirdFunc3 mock[func(a0 map[stub.T]func(in ...*chan<- time.Time)) stub.T]
}

func NewStubInterface[T comparable](t *testing.T) *StubInterface[T] {
	return &StubInterface{
		setupWeirdFunc1: newMock[func(a any, b interface {
			A() int
		})](t),
		setupWeirdFunc2: newMock[func(in *<-chan time.Time, outs ...chan int) error](t),
		setupEmpty:      newMock[func()](t),
		setupWeirdFunc3: newMock[func(a0 map[stub.T]func(in ...*chan<- time.Time)) stub.T](t),
	}
}

func (s *StubInterface) AssertExpectations(t *testing.T) bool {
	return s.setupWeirdFunc1.AssertExpectations(t) &&
		s.setupWeirdFunc2.AssertExpectations(t) &&
		s.setupEmpty.AssertExpectations(t) &&
		s.setupWeirdFunc3.AssertExpectations(t) &&
		true
}

func (s *StubInterface) OnWeirdFunc1(funcs ...func(a any, b interface {
	A() int
})) Config {
	return s.setupWeirdFunc1.append(funcs...)
}

func (s *StubInterface[T]) WeirdFunc1(a any, b interface {
	A() int
}) {
	f, ok := s.setupWeirdFunc1.call()
	if !ok {
		panic(fmt.Sprintf("unexpected call WeirdFunc1(%v,%v)", a, b))
	}
	(*f)(a, b)
}

func (s *StubInterface) OnWeirdFunc2(funcs ...func(in *<-chan time.Time, outs ...chan int) error) Config {
	return s.setupWeirdFunc2.append(funcs...)
}

func (s *StubInterface[T]) WeirdFunc2(in *<-chan time.Time, outs ...chan int) error {
	f, ok := s.setupWeirdFunc2.call()
	if !ok {
		panic(fmt.Sprintf("unexpected call WeirdFunc2(%v,%v)", in, outs))
	}
	return (*f)(in, outs...)
}

func (s *StubInterface) OnEmpty(funcs ...func()) Config {
	return s.setupEmpty.append(funcs...)
}

func (s *StubInterface[T]) Empty() {
	f, ok := s.setupEmpty.call()
	if !ok {
		panic(fmt.Sprintf("unexpected call Empty()"))
	}
	(*f)()
}

func (s *StubInterface) OnWeirdFunc3(funcs ...func(a0 map[stub.T]func(in ...*chan<- time.Time)) stub.T) Config {
	return s.setupWeirdFunc3.append(funcs...)
}

func (s *StubInterface[T]) WeirdFunc3(a0 map[stub.T]func(in ...*chan<- time.Time)) stub.T {
	f, ok := s.setupWeirdFunc3.call()
	if !ok {
		panic(fmt.Sprintf("unexpected call WeirdFunc3(%v)", a0))
	}
	return (*f)(a0)
}