// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// Terraform Plugin Framework variants of standard acceptance test helpers.

// CheckFrameworkResourceDisappears destroys an existing resource out of band
//
// By default, this check will only copy root-level string arguments into the state
// used to delete the remote resource. For resources requiring nested or non-string
// arguments to be available for the delete operation, consider using
// CheckFrameworkResourceDisappearsWithStateFunc with a custom state function
// instead.
func CheckFrameworkResourceDisappears(
	ctx context.Context,
	provo *schema.Provider,
	factory func(context.Context) (fwresource.ResourceWithConfigure, error),
	n string,
) resource.TestCheckFunc {
	return deleteFrameworkResource(ctx, provo, factory, n, rootStringStateFunc())
}

// CheckFrameworkResourceDisappearsWithStateFunc destroys an existing resource
// out of band, constructing state from the provided state function
func CheckFrameworkResourceDisappearsWithStateFunc(
	ctx context.Context,
	provo *schema.Provider,
	factory func(context.Context) (fwresource.ResourceWithConfigure, error),
	n string,
	stateFunc func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error,
) resource.TestCheckFunc {
	return deleteFrameworkResource(ctx, provo, factory, n, stateFunc)
}

func deleteFrameworkResource(
	ctx context.Context,
	provo *schema.Provider,
	factory func(context.Context) (fwresource.ResourceWithConfigure, error),
	n string,
	stateFunc func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", n)
		}

		resource, err := factory(ctx)
		if err != nil {
			return err
		}

		resource.Configure(ctx, fwresource.ConfigureRequest{ProviderData: provo.Meta()}, &fwresource.ConfigureResponse{})

		schemaResp := fwresource.SchemaResponse{}
		resource.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)

		// Construct a simple Framework State that contains just top-level attributes.
		state := tfsdk.State{
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
			Schema: schemaResp.Schema,
		}

		err = stateFunc(ctx, &state, rs.Primary)
		if err != nil {
			return err
		}

		response := fwresource.DeleteResponse{}
		resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

		if response.Diagnostics.HasError() {
			return fwdiag.DiagnosticsError(response.Diagnostics)
		}

		return nil
	}
}

// rootStringStateFunc copies root-level string arguments into `state`
func rootStringStateFunc() func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	return func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
		for name, v := range is.Attributes {
			if name == "%" || strings.Contains(name, ".") {
				continue
			}

			if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root(name), v)); err != nil {
				log.Printf("[WARN] %s(%s): %s", name, v, err)
			}
		}
		return nil
	}
}
