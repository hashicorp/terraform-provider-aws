//go:build sweep
// +build sweep

package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Terraform Plugin Framework variants of sweeper helpers.

type SweepFrameworkResource struct {
	factory func(context.Context) (fwresource.ResourceWithConfigure, error)
	id      string // TODO Currently we can only delete a resource if "id" is the only attribute used.
	meta    interface{}
}

func NewSweepFrameworkResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), id string, meta interface{}) *SweepFrameworkResource {
	return &SweepFrameworkResource{
		factory: factory,
		id:      id,
		meta:    meta,
	}
}

func (sr *SweepFrameworkResource) Delete(ctx context.Context, rc RetryConfig) error {
	err := tfresource.RetryConfigContext(ctx, rc.Delay, rc.DelayRand, rc.MinTimeout, rc.PollInterval, rc.Timeout, func() *resource.RetryError {
		err := DeleteFrameworkResource(sr.factory, sr.id, sr.meta)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[INFO] While sweeping resource (%s), encountered throttling error (%s). Retrying...", sr.id, err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		err = DeleteFrameworkResource(sr.factory, sr.id, sr.meta)
	}

	return err

}

func DeleteFrameworkResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), id string, meta interface{}) error {
	ctx := context.Background()

	resource, err := factory(ctx)

	if err != nil {
		return err
	}

	resource.Configure(ctx, fwresource.ConfigureRequest{ProviderData: meta}, &fwresource.ConfigureResponse{})

	schema, diags := resource.GetSchema(ctx)

	if diags.HasError() {
		return errs.NewDiagnosticsError(diags)
	}

	// Simple Terraform State that contains just the resource ID.
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}
	state.SetAttribute(ctx, path.Root("id"), id)
	response := fwresource.DeleteResponse{}
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	if response.Diagnostics.HasError() {
		return errs.NewDiagnosticsError(response.Diagnostics)
	}

	return nil
}
