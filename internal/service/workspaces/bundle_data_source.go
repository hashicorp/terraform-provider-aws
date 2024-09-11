// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_bundle")
func DataSourceBundle() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceBundleRead,

		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrOwner, names.AttrName},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"bundle_id"},
			},
			names.AttrOwner: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"bundle_id"},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_type": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"root_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWorkspaceBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	var bundle types.WorkspaceBundle

	if bundleID, ok := d.GetOk("bundle_id"); ok {
		resp, err := conn.DescribeWorkspaceBundles(ctx, &workspaces.DescribeWorkspaceBundlesInput{
			BundleIds: []string{bundleID.(string)},
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WorkSpaces Workspace Bundle (%s): %s", bundleID, err)
		}

		if len(resp.Bundles) != 1 {
			return sdkdiag.AppendErrorf(diags, "expected 1 result for WorkSpaces Workspace Bundle %q, found %d", bundleID, len(resp.Bundles))
		}

		if len(resp.Bundles) == 0 {
			return sdkdiag.AppendErrorf(diags, "no WorkSpaces Workspace Bundle with ID %q found", bundleID)
		}

		bundle = resp.Bundles[0]
	}

	if name, ok := d.GetOk(names.AttrName); ok {
		id := name
		input := &workspaces.DescribeWorkspaceBundlesInput{}

		if owner, ok := d.GetOk(names.AttrOwner); ok {
			id = fmt.Sprintf("%s:%s", owner, id)
			input.Owner = aws.String(owner.(string))
		}

		name := name.(string)

		paginator := workspaces.NewDescribeWorkspaceBundlesPaginator(conn, input, func(out *workspaces.DescribeWorkspaceBundlesPaginatorOptions) {})

		entryNotFound := true
		for paginator.HasMorePages() && entryNotFound {
			out, err := paginator.NextPage(ctx)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading WorkSpaces Workspace Bundle (%s): %s", id, err)
			}

			for _, b := range out.Bundles {
				if aws.ToString(b.Name) == name {
					bundle = b
					entryNotFound = false
				}
			}
		}

		if entryNotFound {
			return sdkdiag.AppendErrorf(diags, "no WorkSpaces Workspace Bundle with name %q found", name)
		}
	}

	d.SetId(aws.ToString(bundle.BundleId))
	d.Set("bundle_id", bundle.BundleId)
	d.Set(names.AttrDescription, bundle.Description)
	d.Set(names.AttrName, bundle.Name)
	d.Set(names.AttrOwner, bundle.Owner)

	computeType := make([]map[string]interface{}, 1)
	if bundle.ComputeType != nil {
		computeType[0] = map[string]interface{}{
			names.AttrName: string(bundle.ComputeType.Name),
		}
	}
	if err := d.Set("compute_type", computeType); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_type: %s", err)
	}

	rootStorage := make([]map[string]interface{}, 1)
	if bundle.RootStorage != nil {
		rootStorage[0] = map[string]interface{}{
			"capacity": aws.ToString(bundle.RootStorage.Capacity),
		}
	}
	if err := d.Set("root_storage", rootStorage); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_storage: %s", err)
	}

	userStorage := make([]map[string]interface{}, 1)
	if bundle.UserStorage != nil {
		userStorage[0] = map[string]interface{}{
			"capacity": aws.ToString(bundle.UserStorage.Capacity),
		}
	}
	if err := d.Set("user_storage", userStorage); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user_storage: %s", err)
	}

	return diags
}
