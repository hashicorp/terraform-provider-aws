package ram

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ram_sharing_with_organization")
func ResourceSharingWithOrganization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSharingWithOrganizationCreate,
		ReadWithoutTimeout:   schema.NoopContext,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{},
	}
}

func resourceSharingWithOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn()

	log.Print("[DEBUG] Enabling RAM sharing with organization")

	resp, err := conn.EnableSharingWithAwsOrganizationWithContext(ctx, &ram.EnableSharingWithAwsOrganizationInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error enabling RAM sharing with organization: %s", err)
	}

	if !aws.BoolValue(resp.ReturnValue) {
		return sdkdiag.AppendErrorf(diags, "RAM sharing with organization is not enabled")
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return diags
}
