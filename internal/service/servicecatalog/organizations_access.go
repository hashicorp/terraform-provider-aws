// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_organizations_access")
func ResourceOrganizationsAccess() *schema.Resource {
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

func resourceOrganizationsAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	d.SetId(meta.(*conns.AWSClient).AccountID)

	// During create, if enabled = "true", then Enable Access and vice versa
	// During delete, the opposite

	if _, ok := d.GetOk(names.AttrEnabled); ok {
		_, err := conn.EnableAWSOrganizationsAccessWithContext(ctx, &servicecatalog.EnableAWSOrganizationsAccessInput{})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Service Catalog AWS Organizations Access: %s", err)
		}

		return append(diags, resourceOrganizationsAccessRead(ctx, d, meta)...)
	}

	_, err := conn.DisableAWSOrganizationsAccessWithContext(ctx, &servicecatalog.DisableAWSOrganizationsAccessInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Service Catalog AWS Organizations Access: %s", err)
	}

	return append(diags, resourceOrganizationsAccessRead(ctx, d, meta)...)
}

func resourceOrganizationsAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	output, err := WaitOrganizationsAccessStable(ctx, conn, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
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

	if output == servicecatalog.AccessStatusEnabled {
		d.Set(names.AttrEnabled, true)
		return diags
	}

	d.Set(names.AttrEnabled, false)
	return diags
}

func resourceOrganizationsAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	// During create, if enabled = "true", then Enable Access and vice versa
	// During delete, the opposite

	if _, ok := d.GetOk(names.AttrEnabled); !ok {
		_, err := conn.EnableAWSOrganizationsAccessWithContext(ctx, &servicecatalog.EnableAWSOrganizationsAccessInput{})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Service Catalog AWS Organizations Access: %s", err)
		}

		return diags
	}

	_, err := conn.DisableAWSOrganizationsAccessWithContext(ctx, &servicecatalog.DisableAWSOrganizationsAccessInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Service Catalog AWS Organizations Access: %s", err)
	}

	return diags
}
