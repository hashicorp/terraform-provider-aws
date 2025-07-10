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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_licensemanager_grant_accepter", name="Grant Accepter")
func resourceGrantAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGrantAccepterCreate,
		ReadWithoutTimeout:   resourceGrantAccepterRead,
		DeleteWithoutTimeout: resourceGrantAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allowed_operations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Allowed operations for the grant.",
			},
			"grant_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				Description:  "Amazon Resource Name (ARN) of the grant.",
			},
			"home_region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Home Region of the grant.",
			},
			"license_arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "License ARN.",
			},
			names.AttrName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the grant.",
			},
			"parent_arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent ARN.",
			},
			names.AttrPrincipal: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The grantee principal ARN.",
			},
			names.AttrStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "GrantAccepter status.",
			},
			names.AttrVersion: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "GrantAccepter version.",
			},
		},
	}
}

func resourceGrantAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	grantARN := d.Get("grant_arn").(string)
	input := &licensemanager.AcceptGrantInput{
		GrantArn: aws.String(grantARN),
	}

	output, err := conn.AcceptGrant(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting License Manager Grant (%s): %s", grantARN, err)
	}

	d.SetId(aws.ToString(output.GrantArn))

	return append(diags, resourceGrantAccepterRead(ctx, d, meta)...)
}

func resourceGrantAccepterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	grant, err := findReceivedGrantByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] License Manager Received Grant %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Received Grant (%s): %s", d.Id(), err)
	}

	d.Set("allowed_operations", grant.GrantedOperations)
	d.Set("grant_arn", grant.GrantArn)
	d.Set("home_region", grant.HomeRegion)
	d.Set("license_arn", grant.LicenseArn)
	d.Set(names.AttrName, grant.GrantName)
	d.Set("parent_arn", grant.ParentArn)
	d.Set(names.AttrPrincipal, grant.GranteePrincipalArn)
	d.Set(names.AttrStatus, grant.GrantStatus)
	d.Set(names.AttrVersion, grant.Version)

	return diags
}

func resourceGrantAccepterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	_, err := conn.RejectGrant(ctx, &licensemanager.RejectGrantInput{
		GrantArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "rejecting License Manager Grant (%s): %s", d.Id(), err)
	}

	return diags
}

func findReceivedGrantByARN(ctx context.Context, conn *licensemanager.Client, arn string) (*awstypes.Grant, error) {
	input := &licensemanager.ListReceivedGrantsInput{
		GrantArns: []string{arn},
	}

	return findReceivedGrant(ctx, conn, input, func(v *awstypes.Grant) bool {
		if aws.ToString(v.GrantArn) == arn {
			if status := v.GrantStatus; status == awstypes.GrantStatusActive || status == awstypes.GrantStatusDisabled {
				return true
			}
		}

		return false
	})
}

func findReceivedGrant(ctx context.Context, conn *licensemanager.Client, input *licensemanager.ListReceivedGrantsInput, filter tfslices.Predicate[*awstypes.Grant]) (*awstypes.Grant, error) {
	output, err := findReceivedGrants(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReceivedGrants(ctx context.Context, conn *licensemanager.Client, input *licensemanager.ListReceivedGrantsInput, filter tfslices.Predicate[*awstypes.Grant]) ([]awstypes.Grant, error) {
	var output []awstypes.Grant

	err := listReceivedGrantsPages(ctx, conn, input, func(page *licensemanager.ListReceivedGrantsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Grants {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, err
}
