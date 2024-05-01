// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

var (
	_ WithMeta = (*withMeta)(nil)
)

type WithMeta interface {
	Meta() *conns.AWSClient
}

type withMeta struct {
	meta *conns.AWSClient
}

// Meta returns the provider Meta (instance data).
func (w *withMeta) Meta() *conns.AWSClient {
	return w.meta
}

// RegionalARN returns a regional ARN for the specified service namespace and resource.
func (w *withMeta) RegionalARN(service, resource string) string {
	return arn.ARN{
		Partition: w.meta.Partition,
		Service:   service,
		Region:    w.meta.Region,
		AccountID: w.meta.AccountID,
		Resource:  resource,
	}.String()
}
