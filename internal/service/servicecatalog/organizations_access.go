// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_organizations_access", name="Organizations Access")
func resourceOrganizationsAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationsAccessCreate,
		ReadWithoutTimeout:   resourceOrganizationsAccessRead,
		DeleteWithoutTimeout: resourceOrganizationsAccessDelete,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(OrganizationsAccessStableTimeout),
		},

		Schema: map[string]*schema.Schema{
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceOrganizationsAccessCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	d.SetId(meta.(*conns.AWSClient).AccountID(ctx))

	// During create, if enabled = "true", then Enable Access and vice versa
	// During delete, the opposite

	if _, ok := d.GetOk(names.AttrEnabled); ok {
		_, err := conn.EnableAWSOrganizationsAccess(ctx, &servicecatalog.EnableAWSOrganizationsAccessInput{})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Service Catalog AWS Organizations Access: %s", err)
		}

		return append(diags, resourceOrganizationsAccessRead(ctx, d, meta)...)
	}

	_, err := conn.DisableAWSOrganizationsAccess(ctx, &servicecatalog.DisableAWSOrganizationsAccessInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Service Catalog AWS Organizations Access: %s", err)
	}

	return append(diags, resourceOrganizationsAccessRead(ctx, d, meta)...)
}

func resourceOrganizationsAccessRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	output, err := waitOrganizationsAccessStable(ctx, conn, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		// theoretically this should not be possible
		log.Printf("[WARN] Service Catalog Organizations Access (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog AWS Organizations Access (%s): %s", d.Id(), err)
	}

	if output == "" {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog AWS Organizations Access (%s): empty response", d.Id())
	}

	if output == string(awstypes.AccessStatusEnabled) {
		d.Set(names.AttrEnabled, true)
		return diags
	}

	d.Set(names.AttrEnabled, false)
	return diags
}

func resourceOrganizationsAccessDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	// During create, if enabled = "true", then Enable Access and vice versa
	// During delete, the opposite

	if _, ok := d.GetOk(names.AttrEnabled); !ok {
		_, err := conn.EnableAWSOrganizationsAccess(ctx, &servicecatalog.EnableAWSOrganizationsAccessInput{})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Service Catalog AWS Organizations Access: %s", err)
		}

		return diags
	}

	_, err := conn.DisableAWSOrganizationsAccess(ctx, &servicecatalog.DisableAWSOrganizationsAccessInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Service Catalog AWS Organizations Access: %s", err)
	}

	return diags
}
