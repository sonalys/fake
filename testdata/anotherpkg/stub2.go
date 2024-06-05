package anotherpkg

import "context"

type ExternalType int

type ExternalInterface interface {
	A(ExternalType) func(context.Context)
}
