// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_eks_addon_version")
func dataSourceAddonVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAddonVersionRead,

		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"kubernetes_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAddonVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	addonName := d.Get("addon_name").(string)
	kubernetesVersion := d.Get("kubernetes_version").(string)
	mostRecent := d.Get("most_recent").(bool)
	versionInfo, err := findAddonVersionByTwoPartKey(ctx, conn, addonName, kubernetesVersion, mostRecent)

	if err != nil {
		return diag.Errorf("reading EKS Add-On version info (%s, %s): %s", addonName, kubernetesVersion, err)
	}

	d.SetId(addonName)
	d.Set("addon_name", addonName)
	d.Set("kubernetes_version", kubernetesVersion)
	d.Set("most_recent", mostRecent)
	d.Set("version", versionInfo.AddonVersion)

	return nil
}

func findAddonVersionByTwoPartKey(ctx context.Context, conn *eks.Client, addonName, kubernetesVersion string, mostRecent bool) (*types.AddonVersionInfo, error) {
	input := &eks.DescribeAddonVersionsInput{
		AddonName:         aws.String(addonName),
		KubernetesVersion: aws.String(kubernetesVersion),
	}

	pages := eks.NewDescribeAddonVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Addons {
			for i, v := range v.AddonVersions {
				if mostRecent && i == 0 && v.AddonVersion != nil {
					return &v, nil
				}

				for _, compatibility := range v.Compatibilities {
					if compatibility.DefaultVersion && v.AddonVersion != nil {
						return &v, nil
					}
				}
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}
