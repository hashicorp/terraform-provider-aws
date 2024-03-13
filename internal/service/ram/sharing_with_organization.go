// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ram_sharing_with_organization", name="Sharing With Organization")
func resourceSharingWithOrganization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSharingWithOrganizationCreate,
		ReadWithoutTimeout:   resourceSharingWithOrganizationRead,
		DeleteWithoutTimeout: resourceSharingWithOrganizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{},
	}
}

const (
	sharingWithOrganizationRoleName = "AWSServiceRoleForResourceAccessManager"
	servicePrincipalName            = "ram.amazonaws.com"
)

func resourceSharingWithOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	output, err := conn.EnableSharingWithAwsOrganizationWithContext(ctx, &ram.EnableSharingWithAwsOrganizationInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling RAM Sharing With Organization: %s", err)
	}

	if !aws.BoolValue(output.ReturnValue) {
		return sdkdiag.AppendErrorf(diags, "RAM Sharing With Organization failed")
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceSharingWithOrganizationRead(ctx, d, meta)...)
}

func resourceSharingWithOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// See https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html#getting-started-sharing-orgs.
	// Check for IAM role and Organizations trusted access.

	_, err := tfiam.FindRoleByName(ctx, meta.(*conns.AWSClient).IAMConn(ctx), sharingWithOrganizationRoleName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Role (%s) not found, removing from state", sharingWithOrganizationRoleName)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): %s", sharingWithOrganizationRoleName, err)
	}

	servicePrincipalNames, err := tforganizations.FindEnabledServicePrincipalNames(ctx, meta.(*conns.AWSClient).OrganizationsConn(ctx))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organization service principals: %s", err)
	}

	enabled := slices.Contains(servicePrincipalNames, servicePrincipalName)

	if !d.IsNewResource() && !enabled {
		log.Printf("[WARN] Organization service principal (%s) not enabled, removing from state", servicePrincipalName)
		d.SetId("")
		return diags
	}

	return diags
}

func resourceSharingWithOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// See https://docs.aws.amazon.com/ram/latest/userguide/security-disable-sharing-with-orgs.html.

	if err := tforganizations.DisableServicePrincipal(ctx, meta.(*conns.AWSClient).OrganizationsConn(ctx), servicePrincipalName); err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Organization service principal (%s): %s", servicePrincipalName, err)
	}

	if err := tfiam.DeleteServiceLinkedRole(ctx, meta.(*conns.AWSClient).IAMConn(ctx), sharingWithOrganizationRoleName); err != nil {
		return sdkdiag.AppendWarningf(diags, "deleting IAM service-linked Role (%s): %s", sharingWithOrganizationRoleName, err)
	}

	return diags
}
