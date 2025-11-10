// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_licensemanager_association", name="Association")
func resourceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAssociationCreate,
		ReadWithoutTimeout:   resourceAssociationRead,
		DeleteWithoutTimeout: resourceAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"license_configuration_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	licenseConfigurationARN := d.Get("license_configuration_arn").(string)
	resourceARN := d.Get(names.AttrResourceARN).(string)
	id := associationCreateResourceID(resourceARN, licenseConfigurationARN)
	input := &licensemanager.UpdateLicenseSpecificationsForResourceInput{
		AddLicenseSpecifications: []awstypes.LicenseSpecification{{
			LicenseConfigurationArn: aws.String(licenseConfigurationARN),
		}},
		ResourceArn: aws.String(resourceARN),
	}

	_, err := conn.UpdateLicenseSpecificationsForResource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating License Manager Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAssociationRead(ctx, d, meta)...)
}

func resourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	resourceARN, licenseConfigurationARN, err := associationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = findAssociationByTwoPartKey(ctx, conn, resourceARN, licenseConfigurationARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] License Manager Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Association (%s): %s", d.Id(), err)
	}

	d.Set("license_configuration_arn", licenseConfigurationARN)
	d.Set(names.AttrResourceARN, resourceARN)

	return diags
}

func resourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	resourceARN, licenseConfigurationARN, err := associationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.UpdateLicenseSpecificationsForResource(ctx, &licensemanager.UpdateLicenseSpecificationsForResourceInput{
		RemoveLicenseSpecifications: []awstypes.LicenseSpecification{{
			LicenseConfigurationArn: aws.String(licenseConfigurationARN),
		}},
		ResourceArn: aws.String(resourceARN),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting License Manager Association (%s): %s", d.Id(), err)
	}

	return diags
}

const associationResourceIDSeparator = ","

func associationCreateResourceID(resourceARN, licenseConfigurationARN string) string {
	parts := []string{resourceARN, licenseConfigurationARN}
	id := strings.Join(parts, associationResourceIDSeparator)

	return id
}

func associationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, associationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ResourceARN%[2]sLicenseConfigurationARN", id, associationResourceIDSeparator)
}

func findAssociationByTwoPartKey(ctx context.Context, conn *licensemanager.Client, resourceARN, licenseConfigurationARN string) error {
	input := &licensemanager.ListLicenseSpecificationsForResourceInput{
		ResourceArn: aws.String(resourceARN),
	}

	_, err := findLicenseSpecification(ctx, conn, input, func(v *awstypes.LicenseSpecification) bool {
		return aws.ToString(v.LicenseConfigurationArn) == licenseConfigurationARN
	})

	return err
}

func findLicenseSpecification(ctx context.Context, conn *licensemanager.Client, input *licensemanager.ListLicenseSpecificationsForResourceInput, filter tfslices.Predicate[*awstypes.LicenseSpecification]) (*awstypes.LicenseSpecification, error) {
	output, err := findLicenseSpecifications(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLicenseSpecifications(ctx context.Context, conn *licensemanager.Client, input *licensemanager.ListLicenseSpecificationsForResourceInput, filter tfslices.Predicate[*awstypes.LicenseSpecification]) ([]awstypes.LicenseSpecification, error) {
	var output []awstypes.LicenseSpecification

	err := listLicenseSpecificationsForResourcePages(ctx, conn, input, func(page *licensemanager.ListLicenseSpecificationsForResourceOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LicenseSpecifications {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, err
}
