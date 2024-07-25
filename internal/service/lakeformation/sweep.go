// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	sweep.Register("aws_lakeformation_permissions", sweepPermissions)

	sweep.Register("aws_lakeformation_resource", sweepResource)
}

func sweepPermissions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LakeFormationClient(ctx)

	var sweepResources []sweep.Sweepable
	r := ResourcePermissions()

	pages := lakeformation.NewListPermissionsPaginator(conn, &lakeformation.ListPermissionsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.PrincipalResourcePermissions {
			d := r.Data(nil)

			d.Set(names.AttrPrincipal, v.Principal.DataLakePrincipalIdentifier)
			d.Set(names.AttrPermissions, flattenResourcePermissions([]types.PrincipalResourcePermissions{v}))
			d.Set("permissions_with_grant_option", flattenGrantPermissions([]types.PrincipalResourcePermissions{v}))

			d.Set("catalog_resource", v.Resource.Catalog != nil)

			if v.Resource.DataLocation != nil {
				d.Set("data_location", []any{flattenDataLocationResource(v.Resource.DataLocation)})
			}

			if v.Resource.Database != nil {
				d.Set(names.AttrDatabase, []any{flattenDatabaseResource(v.Resource.Database)})
			}

			if v.Resource.DataCellsFilter != nil {
				d.Set("data_cells_filter", flattenDataCellsFilter(v.Resource.DataCellsFilter))
			}

			if v.Resource.LFTag != nil {
				d.Set("lf_tag", []any{flattenLFTagKeyResource(v.Resource.LFTag)})
			}

			if v.Resource.LFTagPolicy != nil {
				d.Set("lf_tag_policy", []any{flattenLFTagPolicyResource(v.Resource.LFTagPolicy)})
			}

			if v.Resource.Table != nil {
				d.Set("table", []any{flattenTableResource(v.Resource.Table)})
			}

			if v.Resource.TableWithColumns != nil {
				d.Set("table_with_columns", []any{flattenTableColumnsResource(v.Resource.TableWithColumns)})
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepResource(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LakeFormationClient(ctx)

	var sweepResources []sweep.Sweepable
	r := ResourceResource()

	pages := lakeformation.NewListResourcesPaginator(conn, &lakeformation.ListResourcesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResourceInfoList {
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ResourceArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
