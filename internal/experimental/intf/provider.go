package intf

import (
	"context"
)

type ProviderData interface {
	Services(context.Context) map[string]ServiceData
}
