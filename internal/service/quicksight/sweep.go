// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_quicksight_dashboard", &resource.Sweeper{
		Name: "aws_quicksight_dashboard",
		F:    sweepDashboards,
	})
	resource.AddTestSweepers("aws_quicksight_data_set", &resource.Sweeper{
		Name: "aws_quicksight_data_set",
		F:    sweepDataSets,
	})
	resource.AddTestSweepers("aws_quicksight_data_source", &resource.Sweeper{
		Name: "aws_quicksight_data_source",
		F:    sweepDataSources,
	})
	resource.AddTestSweepers("aws_quicksight_folder", &resource.Sweeper{
		Name: "aws_quicksight_folder",
		F:    sweepFolders,
	})
	resource.AddTestSweepers("aws_quicksight_group", &resource.Sweeper{
		Name: "aws_quicksight_group",
		F:    sweepGroups,
	})
	resource.AddTestSweepers("aws_quicksight_template", &resource.Sweeper{
		Name: "aws_quicksight_template",
		F:    sweepTemplates,
	})
	resource.AddTestSweepers("aws_quicksight_user", &resource.Sweeper{
		Name: "aws_quicksight_user",
		F:    sweepUsers,
		Dependencies: []string{
			"aws_quicksight_group",
		},
	})
	resource.AddTestSweepers("aws_quicksight_vpc_connection", &resource.Sweeper{
		Name: "aws_quicksight_vpc_connection",
		F:    sweepVPCConnections,
	})
}

const (
	// Defined locally to avoid cyclic import from internal/acctest
	acctestResourcePrefix = "tf-acc-test"
)

func sweepDashboards(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	awsAccountID := client.AccountID(ctx)
	input := &quicksight.ListDashboardsInput{
		AwsAccountId: aws.String(awsAccountID),
	}

	pages := quicksight.NewListDashboardsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepError(err) {
			log.Printf("[WARN] Skipping QuickSight Dashboard sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Dashboards (%s): %w", region, err)
		}

		for _, v := range page.DashboardSummaryList {
			r := resourceDashboard()
			d := r.Data(nil)
			d.SetId(dashboardCreateResourceID(awsAccountID, aws.ToString(v.DashboardId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight Dashboards (%s): %w", region, err)
	}

	return nil
}

func sweepDataSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	awsAccountID := client.AccountID(ctx)
	input := &quicksight.ListDataSetsInput{
		AwsAccountId: aws.String(awsAccountID),
	}

	pages := quicksight.NewListDataSetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepError(err) {
			log.Printf("[WARN] Skipping QuickSight Data Set sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Data Sets (%s): %w", region, err)
		}

		for _, v := range page.DataSetSummaries {
			r := resourceDataSet()
			d := r.Data(nil)
			d.SetId(dataSetCreateResourceID(awsAccountID, aws.ToString(v.DataSetId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight Data Sets (%s): %w", region, err)
	}

	return nil
}

func sweepDataSources(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	awsAccountID := client.AccountID(ctx)
	input := &quicksight.ListDataSourcesInput{
		AwsAccountId: aws.String(awsAccountID),
	}

	pages := quicksight.NewListDataSourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepError(err) {
			log.Printf("[WARN] Skipping QuickSight Data Source sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Data Sources (%s): %w", region, err)
		}

		for _, v := range page.DataSources {
			r := resourceDataSource()
			d := r.Data(nil)
			d.SetId(dataSourceCreateResourceID(awsAccountID, aws.ToString(v.DataSourceId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight Data Sources (%s): %w", region, err)
	}

	return nil
}

func sweepFolders(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	accountID := client.AccountID(ctx)
	input := &quicksight.ListFoldersInput{
		AwsAccountId: aws.String(accountID),
	}

	pages := quicksight.NewListFoldersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepError(err) {
			log.Printf("[WARN] Skipping QuickSight Folder sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Folders (%s): %w", region, err)
		}

		for _, v := range page.FolderSummaryList {
			r := resourceFolder()
			d := r.Data(nil)
			d.SetId(folderCreateResourceID(accountID, aws.ToString(v.FolderId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight Folders (%s): %w", region, err)
	}

	return nil
}

func sweepGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	awsAccountID := client.AccountID(ctx)
	input := &quicksight.ListGroupsInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(defaultUserNamespace),
	}

	pages := quicksight.NewListGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepUsersOrGroupsError(err) {
			log.Printf("[WARN] Skipping QuickSight Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Groups (%s): %w", region, err)
		}

		for _, v := range page.GroupList {
			groupName := aws.ToString(v.GroupName)

			if !strings.HasPrefix(groupName, acctestResourcePrefix) {
				log.Printf("[INFO] Skipping QuickSight Group %s", groupName)
				continue
			}

			r := resourceGroup()
			d := r.Data(nil)
			d.SetId(groupCreateResourceID(awsAccountID, defaultUserNamespace, groupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight Groups (%s): %w", region, err)
	}

	return nil
}

func sweepTemplates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	awsAccountID := client.AccountID(ctx)
	input := &quicksight.ListTemplatesInput{
		AwsAccountId: aws.String(awsAccountID),
	}

	pages := quicksight.NewListTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepError(err) {
			log.Printf("[WARN] Skipping QuickSight Template sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Templates (%s): %w", region, err)
		}

		for _, v := range page.TemplateSummaryList {
			r := resourceTemplate()
			d := r.Data(nil)
			d.SetId(templateCreateResourceID(awsAccountID, aws.ToString(v.TemplateId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight Templates (%s): %w", region, err)
	}

	return nil
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	awsAccountID := client.AccountID(ctx)
	input := &quicksight.ListUsersInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(defaultUserNamespace),
	}

	pages := quicksight.NewListUsersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepUsersOrGroupsError(err) {
			log.Printf("[WARN] Skipping QuickSight User sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Users (%s): %w", region, err)
		}

		for _, v := range page.UserList {
			userName := aws.ToString(v.UserName)

			if !strings.HasPrefix(userName, acctestResourcePrefix) {
				log.Printf("[INFO] Skipping QuickSight User %s", userName)
				continue
			}

			r := resourceUser()
			d := r.Data(nil)
			d.SetId(userCreateResourceID(awsAccountID, defaultUserNamespace, userName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight Users (%s): %w", region, err)
	}

	return nil
}

func sweepVPCConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	awsAccountID := client.AccountID(ctx)
	input := &quicksight.ListVPCConnectionsInput{
		AwsAccountId: aws.String(awsAccountID),
	}

	pages := quicksight.NewListVPCConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if skipSweepError(err) {
			log.Printf("[WARN] Skipping QuickSight VPC Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight VPC Connections (%s): %w", region, err)
		}

		for _, v := range page.VPCConnectionSummaries {
			vpcConnectionID := aws.ToString(v.VPCConnectionId)

			if status := v.Status; status == awstypes.VPCConnectionResourceStatusDeleted || status == awstypes.VPCConnectionResourceStatusDeletionFailed {
				log.Printf("[INFO] Skipping QuickSight Group %s: Status=%s", vpcConnectionID, status)
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newVPCConnectionResource, client,
				framework.NewAttribute(names.AttrID, vpcConnectionCreateResourceID(awsAccountID, vpcConnectionID)),
				framework.NewAttribute(names.AttrAWSAccountID, awsAccountID),
				framework.NewAttribute("vpc_connection_id", vpcConnectionID),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QuickSight VPC Connections (%s): %w", region, err)
	}

	return nil
}

func skipSweepError(err error) bool {
	if tfawserr.ErrCodeContains(err, "UnsupportedUserEditionException") {
		return true
	}

	if tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "Directory information for account") ||
		tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "Account information for account") ||
		tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "is not signed up with QuickSight") {
		return true
	}

	return awsv2.SkipSweepError(err)
}

func skipSweepUsersOrGroupsError(err error) bool {
	if tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "is not signed up with QuickSight") ||
		tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "Namespace default not found in account") {
		return true
	}

	return awsv2.SkipSweepError(err)
}
