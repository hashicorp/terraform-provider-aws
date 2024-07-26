// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_organization_admin_account", name="IPAM Organization Admin Account")
func resourceIPAMOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMOrganizationAdminAccountCreate,
		ReadWithoutTimeout:   resourceIPAMOrganizationAdminAccountRead,
		DeleteWithoutTimeout: resourceIPAMOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delegated_admin_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_principal": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	ipamServicePrincipal = "ipam.amazonaws.com"
)

func resourceIPAMOrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	adminAccountID := d.Get("delegated_admin_account_id").(string)
	input := &ec2.EnableIpamOrganizationAdminAccountInput{
		DelegatedAdminAccountId: aws.String(adminAccountID),
	}

	output, err := conn.EnableIpamOrganizationAdminAccount(ctx, input)

	if err == nil && !aws.ToBool(output.Success) {
		err = errors.New("failed")
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling IPAM Organization Admin Account (%s): %s", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	return append(diags, resourceIPAMOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceIPAMOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	account, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, d.Id(), ipamServicePrincipal)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Organization Admin Account: (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, account.Arn)
	d.Set("delegated_admin_account_id", account.Id)
	d.Set(names.AttrEmail, account.Email)
	d.Set(names.AttrName, account.Name)
	d.Set("service_principal", ipamServicePrincipal)

	return diags
}

func resourceIPAMOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting IPAM Organization Admin Account: %s", d.Id())
	output, err := conn.DisableIpamOrganizationAdminAccount(ctx, &ec2.DisableIpamOrganizationAdminAccountInput{
		DelegatedAdminAccountId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeIPAMOrganizationAccountNotRegistered) {
		return diags
	}

	if err == nil && !aws.ToBool(output.Success) {
		err = errors.New("failed")
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling IPAM Organization Admin Account (%s): %s", d.Id(), err)
	}

	return diags
}
