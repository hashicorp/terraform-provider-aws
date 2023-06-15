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

func NewIDAttribute(value any) attribute {
	return NewAttribute("id", value)
}

type SweepResource struct {
	factory func(context.Context) (fwresource.ResourceWithConfigure, error)
	id      string
	meta    *conns.AWSClient

	// attributes stores additional attributes to set in state.
	//
	// This can be used in situations where the Delete method requires multiple attributes
	// to destroy the underlying resource.
	attributes []attribute
}

func NewSweepResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), id string, meta *conns.AWSClient, attributes ...attribute) *SweepResource {
	return &SweepResource{
		factory:    factory,
		id:         id,
		meta:       meta,
		attributes: attributes,
	}
}

func (sr *SweepResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		err := deleteResource(ctx, sr.factory, sr.id, sr.meta, sr.attributes)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[INFO] While sweeping resource (%s), encountered throttling error (%s). Retrying...", sr.id, err)
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	}, optFns...)

	if tfresource.TimedOut(err) {
		err = deleteResource(ctx, sr.factory, sr.id, sr.meta, sr.attributes)
	}

	return err
}

func deleteResource(ctx context.Context, factory func(context.Context) (fwresource.ResourceWithConfigure, error), id string, meta *conns.AWSClient, attributes []attribute) error {
	resource, err := factory(ctx)

	if err != nil {
		return err
	}

	resource.Configure(ctx, fwresource.ConfigureRequest{ProviderData: meta}, &fwresource.ConfigureResponse{})

	schemaResp := fwresource.SchemaResponse{}
	resource.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)

	// Simple Terraform State that contains just the resource ID.
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		Schema: schemaResp.Schema,
	}
	if id != "" {
		attributes = append(attributes, NewIDAttribute(id))
	}

	// Set supplemental attributes, if provided
	for _, attr := range attributes {
		d := state.SetAttribute(ctx, path.Root(attr.path), attr.value)
		if d.HasError() {
			return fwdiag.DiagnosticsError(d)
		}
	}

	response := fwresource.DeleteResponse{}
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	return fwdiag.DiagnosticsError(response.Diagnostics)
}
