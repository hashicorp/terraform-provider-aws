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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// Terraform Plugin Framework variants of standard acceptance test helpers.

func DeleteFrameworkResource(factory func(context.Context) (fwresource.ResourceWithConfigure, error), is *terraform.InstanceState, meta interface{}) error {
	ctx := context.Background()

	resource, err := factory(ctx)

	if err != nil {
		return err
	}

	resource.Configure(ctx, fwresource.ConfigureRequest{ProviderData: meta}, &fwresource.ConfigureResponse{})

	schema, diags := resource.GetSchema(ctx)

	if diags.HasError() {
		return fwdiag.DiagnosticsError(diags)
	}

	// Construct a simple Framework State that contains just top-level attributes.
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}

	for name, v := range is.Attributes {
		if strings.Contains(name, ".") {
			continue
		}

		if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root(name), v)); err != nil {
			log.Printf("[WARN] %s(%s): %s", name, v, err)
		}
	}

	response := fwresource.DeleteResponse{}
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)

	if response.Diagnostics.HasError() {
		return fwdiag.DiagnosticsError(response.Diagnostics)
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

		return DeleteFrameworkResource(factory, rs.Primary, provo.Meta())
	}
}
