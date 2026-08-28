// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_odb_exadb_vm_cluster", sweepExaDBVMClusters)
	awsv2.Register("aws_odb_exascale_db_storage_vault", sweepExascaleDBStorageVaults)
}

func sweepExaDBVMClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := odb.ListExadbVmClustersInput{}
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListExadbVmClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ExadbVmClusters {
			sweepResources = append(sweepResources, framework.NewSweepResource(newExaDBVMClusterResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.ExadbVmClusterId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepExascaleDBStorageVaults(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := odb.ListExascaleDbStorageVaultsInput{}
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListExascaleDbStorageVaultsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ExascaleDbStorageVaults {
			sweepResources = append(sweepResources, framework.NewSweepResource(newExascaleDBStorageVaultResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.ExascaleDbStorageVaultId))),
			)
		}
	}

	return sweepResources, nil
}
