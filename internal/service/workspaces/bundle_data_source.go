// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_bundle", name="Bundle")
func dataSourceBundle() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceBundleRead,

		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrOwner, names.AttrName},
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
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
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
		},
	}
}

func dataSourceWorkspaceBundleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	var bundle *types.WorkspaceBundle
	var err error

	if v, ok := d.GetOk("bundle_id"); ok {
		bundle, err = findBundleByID(ctx, conn, v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		input := &workspaces.DescribeWorkspaceBundlesInput{}
		if v, ok := d.GetOk(names.AttrOwner); ok {
			input.Owner = aws.String(v.(string))
		}

		bundle, err = findBundle(ctx, conn, input, func(v *types.WorkspaceBundle) bool {
			return aws.ToString(v.Name) == name
		})
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("WorkSpaces Bundle", err))
	}

	d.SetId(aws.ToString(bundle.BundleId))
	d.Set("bundle_id", bundle.BundleId)
	tfMap := make([]map[string]any, 1)
	if bundle.ComputeType != nil {
		tfMap[0] = map[string]any{
			names.AttrName: string(bundle.ComputeType.Name),
		}
	}
	if err := d.Set("compute_type", tfMap); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_type: %s", err)
	}
	d.Set(names.AttrDescription, bundle.Description)
	d.Set(names.AttrName, bundle.Name)
	d.Set(names.AttrOwner, bundle.Owner)
	tfMap = make([]map[string]any, 1)
	if bundle.RootStorage != nil {
		tfMap[0] = map[string]any{
			"capacity": aws.ToString(bundle.RootStorage.Capacity),
		}
	}
	if err := d.Set("root_storage", tfMap); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_storage: %s", err)
	}
	tfMap = make([]map[string]any, 1)
	if bundle.UserStorage != nil {
		tfMap[0] = map[string]any{
			"capacity": aws.ToString(bundle.UserStorage.Capacity),
		}
	}
	if err := d.Set("user_storage", tfMap); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user_storage: %s", err)
	}

	return diags
}

func findBundleByID(ctx context.Context, conn *workspaces.Client, id string) (*types.WorkspaceBundle, error) {
	input := &workspaces.DescribeWorkspaceBundlesInput{
		BundleIds: []string{id},
	}

	output, err := findBundles(ctx, conn, input, tfslices.PredicateTrue[*types.WorkspaceBundle]())

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

// findBundle returns the first bundle that matches the filter.
func findBundle(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspaceBundlesInput, filter tfslices.Predicate[*types.WorkspaceBundle]) (*types.WorkspaceBundle, error) {
	output, err := findBundles(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertFirstValueResult(output)
}

func findBundles(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspaceBundlesInput, filter tfslices.Predicate[*types.WorkspaceBundle]) ([]types.WorkspaceBundle, error) {
	var output []types.WorkspaceBundle

	pages := workspaces.NewDescribeWorkspaceBundlesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Bundles {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
