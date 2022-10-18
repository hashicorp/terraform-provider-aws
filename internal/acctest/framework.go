package acctest

import (
	"context"
	"fmt"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// Terraform Plugin Framework variants of standard acceptance test helpers.

func DeleteFrameworkResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), id string, meta interface{}) error {
	ctx := context.Background()

	resource, err := factory(ctx)

	if err != nil {
		return err
	}

	resource.Configure(ctx, fwresource.ConfigureRequest{ProviderData: meta}, &fwresource.ConfigureResponse{})

	// Simple Terraform State that contains just the resource ID.
	state := tfsdk.State{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"id": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"id": tftypes.NewValue(tftypes.String, id),
		}),
		Schema: tfsdk.Schema{
			Attributes: map[string]tfsdk.Attribute{
				"id": {
					Type:     types.StringType,
					Optional: true,
					Computed: true,
				},
			},
		},
	}
	response := fwresource.DeleteResponse{}
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	if response.Diagnostics.HasError() {
		return errs.NewDiagnosticsError(response.Diagnostics)
	}

	return nil
}

func CheckFrameworkResourceDisappears(provo *schema.Provider, factory func(context.Context) (fwresource.ResourceWithConfigure, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", n)
		}

		return DeleteFrameworkResource(factory, rs.Primary.ID, provo.Meta())
	}
}
