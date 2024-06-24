// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResGrant = "Grant"
)

// @SDKResource("aws_licensemanager_grant")
func ResourceGrant() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGrantCreate,
		ReadWithoutTimeout:   resourceGrantRead,
		UpdateWithoutTimeout: resourceGrantUpdate,
		DeleteWithoutTimeout: resourceGrantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allowed_operations": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: len(licensemanager.AllowedOperation_Values()),
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(licensemanager.AllowedOperation_Values(), true),
				},
				Description: "Allowed operations for the grant. This is a subset of the allowed operations on the license.",
			},
			names.AttrARN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Amazon Resource Name (ARN) of the grant.",
			},
			"home_region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Home Region of the grant.",
			},
			"license_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				Description:  "License ARN.",
			},
			names.AttrName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the grant.",
			},
			"parent_arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent ARN.",
			},
			names.AttrPrincipal: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				Description:  "The grantee principal ARN. The target account for the grant in the form of the ARN for an account principal of the root user.",
			},
			names.AttrStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Grant status.",
			},
			names.AttrVersion: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Grant version.",
			},
		},
	}
}

func resourceGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	in := &licensemanager.CreateGrantInput{
		AllowedOperations: aws.StringSlice(expandAllowedOperations(d.Get("allowed_operations").(*schema.Set).List())),
		ClientToken:       aws.String(id.UniqueId()),
		GrantName:         aws.String(d.Get(names.AttrName).(string)),
		HomeRegion:        aws.String(meta.(*conns.AWSClient).Region),
		LicenseArn:        aws.String(d.Get("license_arn").(string)),
		Principals:        aws.StringSlice([]string{d.Get(names.AttrPrincipal).(string)}),
	}

	out, err := conn.CreateGrantWithContext(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionCreating, ResGrant, d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.StringValue(out.GrantArn))

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	out, err := FindGrantByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.LicenseManager, create.ErrActionReading, ResGrant, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionReading, ResGrant, d.Id(), err)
	}

	d.Set("allowed_operations", out.GrantedOperations)
	d.Set(names.AttrARN, out.GrantArn)
	d.Set("home_region", out.HomeRegion)
	d.Set("license_arn", out.LicenseArn)
	d.Set(names.AttrName, out.GrantName)
	d.Set("parent_arn", out.ParentArn)
	d.Set(names.AttrPrincipal, out.GranteePrincipalArn)
	d.Set(names.AttrStatus, out.GrantStatus)
	d.Set(names.AttrVersion, out.Version)

	return diags
}

func resourceGrantUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	in := &licensemanager.CreateGrantVersionInput{
		GrantArn:    aws.String(d.Id()),
		ClientToken: aws.String(id.UniqueId()),
	}

	if d.HasChange("allowed_operations") {
		in.AllowedOperations = aws.StringSlice(d.Get("allowed_operations").([]string))
	}

	if d.HasChange(names.AttrName) {
		in.GrantName = aws.String(d.Get(names.AttrName).(string))
	}

	_, err := conn.CreateGrantVersionWithContext(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionUpdating, ResGrant, d.Id(), err)
	}

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	out, err := FindGrantByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.LicenseManager, create.ErrActionReading, ResGrant, d.Id())
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionReading, ResGrant, d.Id(), err)
	}

	in := &licensemanager.DeleteGrantInput{
		GrantArn: aws.String(d.Id()),
		Version:  out.Version,
	}

	_, err = conn.DeleteGrantWithContext(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionDeleting, ResGrant, d.Id(), err)
	}

	return diags
}

func FindGrantByARN(ctx context.Context, conn *licensemanager.LicenseManager, arn string) (*licensemanager.Grant, error) {
	in := &licensemanager.GetGrantInput{
		GrantArn: aws.String(arn),
	}

	out, err := conn.GetGrantWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, licensemanager.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Grant == nil || aws.StringValue(out.Grant.GrantStatus) == licensemanager.GrantStatusDeleted || aws.StringValue(out.Grant.GrantStatus) == licensemanager.GrantStatusRejected {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Grant, nil
}

func expandAllowedOperations(rawOperations []interface{}) []string {
	if rawOperations == nil {
		return nil
	}

	operations := make([]string, 0, 8)

	for _, item := range rawOperations {
		operations = append(operations, item.(string))
	}

	return operations
}
