// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResGrantAccepter = "Grant Accepter"
)

// @SDKResource("aws_licensemanager_grant_accepter")
func ResourceGrantAccepter() *schema.Resource {
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
			"grant_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				Description:  "Amazon Resource Name (ARN) of the grant.",
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

func resourceGrantAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	in := &licensemanager.AcceptGrantInput{
		GrantArn: aws.String(d.Get("grant_arn").(string)),
	}

	out, err := conn.AcceptGrantWithContext(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionCreating, ResGrantAccepter, d.Get("grant_arn").(string), err)
	}

	d.SetId(aws.StringValue(out.GrantArn))

	return append(diags, resourceGrantAccepterRead(ctx, d, meta)...)
}

func resourceGrantAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	out, err := FindGrantAccepterByGrantARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.LicenseManager, create.ErrActionReading, ResGrantAccepter, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionReading, ResGrantAccepter, d.Id(), err)
	}

	d.Set("allowed_operations", out.GrantedOperations)
	d.Set("grant_arn", out.GrantArn)
	d.Set("home_region", out.HomeRegion)
	d.Set("license_arn", out.LicenseArn)
	d.Set(names.AttrName, out.GrantName)
	d.Set("parent_arn", out.ParentArn)
	d.Set(names.AttrPrincipal, out.GranteePrincipalArn)
	d.Set(names.AttrStatus, out.GrantStatus)
	d.Set(names.AttrVersion, out.Version)

	return diags
}

func resourceGrantAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	in := &licensemanager.RejectGrantInput{
		GrantArn: aws.String(d.Id()),
	}

	_, err := conn.RejectGrantWithContext(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.LicenseManager, create.ErrActionDeleting, ResGrantAccepter, d.Id(), err)
	}

	return diags
}

func FindGrantAccepterByGrantARN(ctx context.Context, conn *licensemanager.LicenseManager, arn string) (*licensemanager.Grant, error) {
	in := &licensemanager.ListReceivedGrantsInput{
		GrantArns: aws.StringSlice([]string{arn}),
	}

	out, err := conn.ListReceivedGrantsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, licensemanager.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	var entry *licensemanager.Grant
	entryExists := false

	for _, grant := range out.Grants {
		if arn == aws.StringValue(grant.GrantArn) && (licensemanager.GrantStatusActive == aws.StringValue(grant.GrantStatus) || licensemanager.GrantStatusDisabled == aws.StringValue(grant.GrantStatus)) {
			entry = grant
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return entry, nil
}
