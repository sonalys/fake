package out

import (
	"testing"

	stub "github.com/sonalys/fake/testdata"
	mocks "github.com/sonalys/fake/testdata/out/testdata"
)

func Test_Mock(t *testing.T) {
	mock := mocks.NewStubInterface[int](t)
	var Stub stub.StubInterface[int] = mock
	mock.OnWeirdFunc1(func(a any, b interface{ A() int }) {
		if a == nil {
			t.Fail()
		}
	})
	Stub.WeirdFunc1(1, nil)
}
