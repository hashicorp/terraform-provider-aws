// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_rds_cluster_instance")
func newClusterInstanceResourceAsListResource() inttypes.ListResourceForSDK {
	l := clusterInstanceListResource{}
	l.SetResourceSchema(resourceClusterInstance())
	return &l
}

var _ list.ListResource = &clusterInstanceListResource{}

type clusterInstanceListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *clusterInstanceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.RDSClient(ctx)

	dbClusters := make(map[string]*types.DBCluster)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &rds.DescribeDBInstancesInput{}
		for item, err := range listDBInstances(ctx, conn, input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}
			if aws.ToString(item.DBClusterIdentifier) == "" {
				continue
			}

			identifier := aws.ToString(item.DBInstanceIdentifier)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrIdentifier), identifier)
			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(identifier)
			rd.Set(names.AttrIdentifier, identifier)

			if _, ok := dbClusters[identifier]; !ok {
				dbc, err := findDBClusterByID(ctx, conn, aws.ToString(item.DBClusterIdentifier))
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
				dbClusters[identifier] = dbc
			}

			if request.IncludeResource {
				if err := resourceClusterInstanceFlatten(ctx, &item, rd, dbClusters[identifier]); err != nil {
					tflog.Error(ctx, "Flattening RDS Cluster Instance", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = identifier
			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}
