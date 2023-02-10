package sweep

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Terraform Plugin Framework variants of sweeper helpers.

type FrameworkSupplementalAttribute struct {
	Path  string
	Value string
}

type SweepFrameworkResource struct {
	factory func(context.Context) (fwresource.ResourceWithConfigure, error)
	id      string
	meta    interface{}

	// supplementalAttributes stores additional attributes to set in state.
	//
	// This can be used in situations where the Delete method requires multiple attributes
	// to destroy the underlying resource.
	supplementalAttributes []FrameworkSupplementalAttribute
}

func NewSweepFrameworkResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), id string, meta interface{}, supplementalAttributes ...FrameworkSupplementalAttribute) *SweepFrameworkResource {
	return &SweepFrameworkResource{
		factory:                factory,
		id:                     id,
		meta:                   meta,
		supplementalAttributes: supplementalAttributes,
	}
}

func (sr *SweepFrameworkResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	err := tfresource.Retry(ctx, timeout, func() *resource.RetryError {
		err := DeleteFrameworkResource(sr.factory, sr.id, sr.meta, sr.supplementalAttributes)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[INFO] While sweeping resource (%s), encountered throttling error (%s). Retrying...", sr.id, err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	}, optFns...)

	if tfresource.TimedOut(err) {
		err = DeleteFrameworkResource(sr.factory, sr.id, sr.meta, sr.supplementalAttributes)
	}

	return err
}

func DeleteFrameworkResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), id string, meta interface{}, supplementalAttributes []FrameworkSupplementalAttribute) error {
	ctx := context.Background()

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
	state.SetAttribute(ctx, path.Root("id"), id)

	// Set supplemental attibutes, if provided
	for _, attr := range supplementalAttributes {
		d := state.SetAttribute(ctx, path.Root(attr.Path), attr.Value)
		if d.HasError() {
			return fwdiag.DiagnosticsError(d)
		}
	}

	response := fwresource.DeleteResponse{}
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	return fwdiag.DiagnosticsError(response.Diagnostics)
}
