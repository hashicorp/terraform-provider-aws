// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// RegisterSweepers registers the ODB resource sweepers.
func RegisterSweepers() {
	awsv2.Register("aws_odb_autonomous_database", sweepAutonomousDatabases)
}

func sweepAutonomousDatabases(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ODBClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := odb.NewListAutonomousDatabasesPaginator(conn, &odb.ListAutonomousDatabasesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, database := range page.AutonomousDatabases {
			id := aws.ToString(database.AutonomousDatabaseId)
			name := aws.ToString(database.DisplayName)
			fields := map[string]any{names.AttrID: id, names.AttrDisplayName: name}
			if id == "" {
				tflog.Warn(ctx, "Skipping ODB Autonomous Database without an ID", fields)
				continue
			}
			if !strings.HasPrefix(name, sweep.ResourcePrefix+"-") {
				tflog.Debug(ctx, "Skipping ODB Autonomous Database without an acceptance test name", fields)
				continue
			}
			if database.Status == odbtypes.AutonomousDatabaseResourceStatusTerminated {
				tflog.Debug(ctx, "Skipping terminated ODB Autonomous Database", fields)
				continue
			}

			tflog.Info(ctx, "Scheduling ODB Autonomous Database for sweeping", fields)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceAutonomousDatabase, client,
				framework.NewAttribute(names.AttrID, id)))
		}
	}

	return sweepResources, nil
}
