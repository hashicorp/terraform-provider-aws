// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_licensemanager_grant", name="Grant")
func resourceGrant() *schema.Resource {
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

func resourceGrantCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &licensemanager.CreateGrantInput{
		AllowedOperations: flex.ExpandStringyValueSet[awstypes.AllowedOperation](d.Get("allowed_operations").(*schema.Set)),
		ClientToken:       aws.String(id.UniqueId()),
		GrantName:         aws.String(name),
		HomeRegion:        aws.String(meta.(*conns.AWSClient).Region(ctx)),
		LicenseArn:        aws.String(d.Get("license_arn").(string)),
		Principals:        []string{d.Get(names.AttrPrincipal).(string)},
	}

	output, err := conn.CreateGrant(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating License Manager Grant (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GrantArn))

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	grant, err := findGrantByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] License Manager Grant %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Grant (%s): %s", d.Id(), err)
	}

	d.Set("allowed_operations", grant.GrantedOperations)
	d.Set(names.AttrARN, grant.GrantArn)
	d.Set("home_region", grant.HomeRegion)
	d.Set("license_arn", grant.LicenseArn)
	d.Set(names.AttrName, grant.GrantName)
	d.Set("parent_arn", grant.ParentArn)
	d.Set(names.AttrPrincipal, grant.GranteePrincipalArn)
	d.Set(names.AttrStatus, grant.GrantStatus)
	d.Set(names.AttrVersion, grant.Version)

	return diags
}

func resourceGrantUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	input := &licensemanager.CreateGrantVersionInput{
		ClientToken: aws.String(id.UniqueId()),
		GrantArn:    aws.String(d.Id()),
	}

	if d.HasChange("allowed_operations") {
		input.AllowedOperations = flex.ExpandStringyValueSet[awstypes.AllowedOperation](d.Get("allowed_operations").(*schema.Set))
	}

	if d.HasChange(names.AttrName) {
		input.GrantName = aws.String(d.Get(names.AttrName).(string))
	}

	_, err := conn.CreateGrantVersion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating License Manager Grant (%s): %s", d.Id(), err)
	}

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	_, err := conn.DeleteGrant(ctx, &licensemanager.DeleteGrantInput{
		GrantArn: aws.String(d.Id()),
		Version:  aws.String(names.AttrVersion),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting License Manager Grant (%s): %s", d.Id(), err)
	}

	return diags
}

func findGrantByARN(ctx context.Context, conn *licensemanager.Client, arn string) (*awstypes.Grant, error) {
	input := &licensemanager.GetGrantInput{
		GrantArn: aws.String(arn),
	}

	output, err := findGrant(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.GrantStatus; status == awstypes.GrantStatusDeleted || status == awstypes.GrantStatusRejected {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findGrant(ctx context.Context, conn *licensemanager.Client, input *licensemanager.GetGrantInput) (*awstypes.Grant, error) {
	output, err := conn.GetGrant(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Grant == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Grant, nil
}
