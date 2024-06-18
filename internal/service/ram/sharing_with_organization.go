// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	output, err := conn.EnableSharingWithAwsOrganization(ctx, &ram.EnableSharingWithAwsOrganizationInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling RAM Sharing With Organization: %s", err)
	}

	if !aws.ToBool(output.ReturnValue) {
		return sdkdiag.AppendErrorf(diags, "RAM Sharing With Organization failed")
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceSharingWithOrganizationRead(ctx, d, meta)...)
}

func resourceSharingWithOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err := findSharingWithOrganization(ctx, meta.(*conns.AWSClient))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RAM Sharing With Organization %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Sharing With Organization (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSharingWithOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// See https://docs.aws.amazon.com/ram/latest/userguide/security-disable-sharing-with-orgs.html.

	if err := tforganizations.DisableServicePrincipal(ctx, meta.(*conns.AWSClient).OrganizationsClient(ctx), servicePrincipalName); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := tfiam.DeleteServiceLinkedRole(ctx, meta.(*conns.AWSClient).IAMClient(ctx), sharingWithOrganizationRoleName); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func findSharingWithOrganization(ctx context.Context, awsClient *conns.AWSClient) error {
	// See https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html#getting-started-sharing-orgs.
	// Check for IAM role and Organizations trusted access.
	_, err := tfiam.FindRoleByName(ctx, awsClient.IAMClient(ctx), sharingWithOrganizationRoleName)

	if err != nil {
		return fmt.Errorf("reading IAM Role (%s): %w", sharingWithOrganizationRoleName, err)
	}

	servicePrincipalNames, err := tforganizations.FindEnabledServicePrincipalNames(ctx, awsClient.OrganizationsClient(ctx))

	if err != nil {
		return fmt.Errorf("reading Organization service principals: %w", err)
	}

	if !slices.Contains(servicePrincipalNames, servicePrincipalName) {
		return &retry.NotFoundError{
			Message: fmt.Sprintf("Organization service principal (%s) not enabled", servicePrincipalName),
		}
	}

	return nil
}
