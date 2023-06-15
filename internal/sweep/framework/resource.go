package framework

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
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

type SweepResource struct {
	factory    func(context.Context) (fwresource.ResourceWithConfigure, error)
	meta       *conns.AWSClient
	attributes []attribute
}

func NewSweepResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), meta *conns.AWSClient, attributes ...attribute) *SweepResource {
	return &SweepResource{
		factory:    factory,
		meta:       meta,
		attributes: attributes,
	}
}

func (sr *SweepResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		err := deleteResource(ctx, sr.factory, sr.meta, sr.attributes)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[INFO] While sweeping resource, encountered throttling error (%s). Retrying...", err)
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	}, optFns...)

	if tfresource.TimedOut(err) {
		err = deleteResource(ctx, sr.factory, sr.meta, sr.attributes)
	}

	return err
}

func deleteResource(ctx context.Context, factory func(context.Context) (fwresource.ResourceWithConfigure, error), meta *conns.AWSClient, attributes []attribute) error {
	resource, err := factory(ctx)

	if err != nil {
		return err
	}

	resource.Configure(ctx, fwresource.ConfigureRequest{ProviderData: meta}, &fwresource.ConfigureResponse{})

	schemaResp := fwresource.SchemaResponse{}
	resource.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		Schema: schemaResp.Schema,
	}

	for _, attr := range attributes {
		d := state.SetAttribute(ctx, path.Root(attr.path), attr.value)
		if d.HasError() {
			return fwdiag.DiagnosticsError(d)
		}
		ctx = tflog.SetField(ctx, attr.path, attr.value)
	}

	response := fwresource.DeleteResponse{}
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	return fwdiag.DiagnosticsError(response.Diagnostics)
}
