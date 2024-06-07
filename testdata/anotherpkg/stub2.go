package anotherpkg

import stub "context"

type ExternalType int

type ExternalInterface interface {
	A(ExternalType) func(stub.Context)
}
