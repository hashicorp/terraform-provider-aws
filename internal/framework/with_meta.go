// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type withMeta struct {
	meta *conns.AWSClient
}

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

// WithImportByID is intended to be embedded in resources which import state via the "id" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByID struct{}

func (w *WithImportByID) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}
