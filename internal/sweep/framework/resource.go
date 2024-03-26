// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"math/rand"
	"strings"
	"time"

	awsretry "github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type attribute struct {
	path  string
	value any
}

func NewAttribute(path string, value any) attribute {
	return attribute{
		path:  path,
		value: value,
	}
}

type sweepResource struct {
	factory    func(context.Context) (fwresource.ResourceWithConfigure, error)
	meta       *conns.AWSClient
	attributes []attribute
}

func NewSweepResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), meta *conns.AWSClient, attributes ...attribute) *sweepResource {
	return &sweepResource{
		factory:    factory,
		meta:       meta,
		attributes: attributes,
	}
}

func (sr *sweepResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	resource, err := sr.factory(ctx)

	if err != nil {
		return err
	}

	metadata := resourceMetadata(ctx, resource)
	ctx = tflog.SetField(ctx, "resource_type", metadata.TypeName)

	resource.Configure(ctx, fwresource.ConfigureRequest{ProviderData: sr.meta}, &fwresource.ConfigureResponse{})

	schemaResp := fwresource.SchemaResponse{}
	resource.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		Schema: schemaResp.Schema,
	}

	for _, attr := range sr.attributes {
		d := state.SetAttribute(ctx, path.Root(attr.path), attr.value)
		if d.HasError() {
			return fwdiag.DiagnosticsError(d)
		}
		ctx = tflog.SetField(ctx, attr.path, attr.value)
	}

	tflog.Info(ctx, "Sweeping resource")

	jitter := time.Duration(rand.Int63n(int64(1*time.Second))) - 1*time.Second/2
	defaultOpts := []tfresource.OptionsFunc{
		tfresource.WithMinPollInterval(2*time.Second + jitter),
	}
	// Put defaults first so subsequent optFns will override them
	optFns = append(defaultOpts, optFns...)

	err = tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		err := deleteResource(ctx, state, resource)

		if err != nil {
			var throttled bool
			for _, code := range maps.Keys(awsretry.DefaultThrottleErrorCodes) {
				// The resource delete operation returns a diag.Diagnostics, so we have to do a
				// string comparison instead of checking the error code of an actual error
				if strings.Contains(err.Error(), code) {
					throttled = true
					break
				}
			}
			if throttled {
				tflog.Info(ctx, "Retrying throttling error", map[string]any{
					"err": err.Error(),
				})
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	}, optFns...)

	if tfresource.TimedOut(err) {
		err = deleteResource(ctx, state, resource)
	}

	return err
}

func deleteResource(ctx context.Context, state tfsdk.State, resource fwresource.Resource) error {
	var response fwresource.DeleteResponse
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	return fwdiag.DiagnosticsError(response.Diagnostics)
}

func resourceMetadata(ctx context.Context, resource fwresource.Resource) fwresource.MetadataResponse {
	var response fwresource.MetadataResponse
	resource.Metadata(ctx, fwresource.MetadataRequest{}, &response)

	return response
}
