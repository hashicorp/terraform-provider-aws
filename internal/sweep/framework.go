//go:build sweep
// +build sweep

package sweep

import (
	"context"
	"errors"
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

func (sr *SweepFrameworkResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	err := tfresource.RetryContext(ctx, timeout, func() *resource.RetryError {
		err := DeleteFrameworkResource(sr.factory, sr.id, sr.meta)

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

	schemaResp := fwresource.SchemaResponse{}
	if v, ok := resource.(fwresource.ResourceWithSchema); ok {
		v.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			return fwdiag.DiagnosticsError(schemaResp.Diagnostics)
		}
	} else {
		return errors.New("resource does not implement Schema method")
	}

	// Simple Terraform State that contains just the resource ID.
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		Schema: schemaResp.Schema,
	}
	state.SetAttribute(ctx, path.Root("id"), id)
	response := fwresource.DeleteResponse{}
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	if response.Diagnostics.HasError() {
		return fwdiag.DiagnosticsError(response.Diagnostics)
	}

	return nil
}
