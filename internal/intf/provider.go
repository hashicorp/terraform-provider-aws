package intf

import (
	"context"
)

type ProviderData interface {
	ServiceData(context.Context) map[string]ServiceData
}
