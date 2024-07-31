// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

// go test -bench=BenchmarkProtoV5ProviderServerFactory -benchtime 1x -benchmem -run=B -v ./internal/provider
func BenchmarkProtoV5ProviderServerFactory(b *testing.B) {
	_, p, err := provider.ProtoV5ProviderServerFactory(context.Background())

	if err != nil {
		b.Fatal(err)
	}

	if b.N == 1 {
		b.Logf("%d resources, %d data sources", len(p.ResourcesMap), len(p.DataSourcesMap))
	}
}
