// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
				MaxItems: len(enum.Values[awstypes.AllowedOperation]()),
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.AllowedOperation](),
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

	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	in := &licensemanager.CreateGrantInput{
		AllowedOperations: expandAllowedOperations(d.Get("allowed_operations").(*schema.Set).List()),
		ClientToken:       aws.String(id.UniqueId()),
		GrantName:         aws.String(d.Get(names.AttrName).(string)),
		HomeRegion:        aws.String(meta.(*conns.AWSClient).Region),
		LicenseArn:        aws.String(d.Get("license_arn").(string)),
		Principals:        []string{d.Get(names.AttrPrincipal).(string)},
	}

	out, err := conn.CreateGrant(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionCreating, ResGrant, d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.ToString(out.GrantArn))

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

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

	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	in := &licensemanager.CreateGrantVersionInput{
		GrantArn:    aws.String(d.Id()),
		ClientToken: aws.String(id.UniqueId()),
	}

	if d.HasChange("allowed_operations") {
		in.AllowedOperations = tfslices.ApplyToAll(d.Get("allowed_operations").([]string), func(v string) awstypes.AllowedOperation {
			return awstypes.AllowedOperation(v)
		})
	}

	if d.HasChange(names.AttrName) {
		in.GrantName = aws.String(d.Get(names.AttrName).(string))
	}

	_, err := conn.CreateGrantVersion(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionUpdating, ResGrant, d.Id(), err)
	}

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

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

	_, err = conn.DeleteGrant(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionDeleting, ResGrant, d.Id(), err)
	}

	return diags
}

func FindGrantByARN(ctx context.Context, conn *licensemanager.Client, arn string) (*awstypes.Grant, error) {
	in := &licensemanager.GetGrantInput{
		GrantArn: aws.String(arn),
	}

	out, err := conn.GetGrant(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Grant == nil || out.Grant.GrantStatus == awstypes.GrantStatusDeleted || out.Grant.GrantStatus == awstypes.GrantStatusRejected {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Grant, nil
}

func expandAllowedOperations(rawOperations []interface{}) []awstypes.AllowedOperation {
	if rawOperations == nil {
		return nil
	}

	operations := make([]awstypes.AllowedOperation, 0, 8)

	for _, item := range rawOperations {
		operations = append(operations, awstypes.AllowedOperation(item.(string)))
	}

	return operations
}
