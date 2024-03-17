// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
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
		return fmt.Errorf("error getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListDashboardsInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	pages := quicksight.NewListDashboardsPaginator(conn, input)

	for pages.HasMorePages() {

		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingQuickSight Dashboard sweep for %s: %s", region, err.Error())
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QuickSight Dashboards (%s): %s", region, err.Error())
		}

		for _, dashboard := range page.DashboardSummaryList {
			r := ResourceDashboard()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.ToString(dashboard.DashboardId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Dashboards for %s: %s", region, err.Error())
	}

	return nil
}

func sweepDataSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListDataSetsInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	pages := quicksight.NewListDataSetsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingQuickSight Datasets sweep for %s: %s", region, err.Error())
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Quicksight Datasets(%s): %s", region, err.Error())
		}
		for _, ds := range page.DataSetSummaries {

			r := ResourceDataSet()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.ToString(ds.DataSetId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Data Sets for %s: %s", region, err.Error())
	}

	return nil
}

func sweepDataSources(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListDataSourcesInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	pages := quicksight.NewListDataSourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingQuickSight Data Sources sweep for %s: %s", region, err.Error())
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Quicksight List Data Sources (%s): %s", region, err.Error())
		}

		for _, ds := range page.DataSources {

			r := ResourceDataSource()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s/%s", awsAccountId, aws.ToString(ds.DataSourceId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Data Sources for %s: %s", region, err.Error())
	}

	return nil
}

func sweepFolders(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	awsAccountId := client.AccountID
	sweepResources := make([]sweep.Sweepable, 0)

	input := &quicksight.ListFoldersInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	out, err := conn.ListFolders(ctx, input)
	for _, folder := range out.FolderSummaryList {
		if folder.FolderId == nil {
			continue
		}

		r := ResourceFolder()
		d := r.Data(nil)
		d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.ToString(folder.FolderId)))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping QuickSight Folder sweep for %s: %s", region, err.Error())
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Folders: %s", err.Error())
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Folders for %s: %s", region, err.Error())
	}

	return nil
}

func sweepGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	awsAccountId := client.AccountID
	sweepResources := make([]sweep.Sweepable, 0)

	input := &quicksight.ListGroupsInput{
		AwsAccountId: aws.String(awsAccountId),
		Namespace:    aws.String(DefaultUserNamespace),
	}

	out, err := conn.ListGroups(ctx, input)
	for _, user := range out.GroupList {
		groupname := aws.ToString(user.GroupName)
		if !strings.HasPrefix(groupname, acctestResourcePrefix) {
			continue
		}

		r := ResourceGroup()
		d := r.Data(nil)
		d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountId, DefaultUserNamespace, groupname))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if skipSweepUserError(err) {
		log.Printf("[WARN] Skipping QuickSight Group sweep for %s: %s", region, err.Error())
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Groups: %s", err.Error())
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Groups for %s: %s", region, err.Error())
	}

	return nil
}

func sweepTemplates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListTemplatesInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	pages := quicksight.NewListTemplatesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] SkippingQuicksight Templates sweep for %s: %s", region, err.Error())
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Quicksight Templates (%s): %s", region, err.Error())
		}

		for _, tmpl := range page.TemplateSummaryList {
			r := ResourceTemplate()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.ToString(tmpl.TemplateId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Templates for %s: %s", region, err.Error())
	}

	return nil
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	awsAccountId := client.AccountID
	sweepResources := make([]sweep.Sweepable, 0)

	input := &quicksight.ListUsersInput{
		AwsAccountId: aws.String(awsAccountId),
		Namespace:    aws.String(DefaultUserNamespace),
	}

	out, err := conn.ListUsers(ctx, input)
	for _, user := range out.UserList {
		username := aws.ToString(user.UserName)
		if !strings.HasPrefix(username, acctestResourcePrefix) {
			continue
		}

		r := ResourceUser()
		d := r.Data(nil)
		d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountId, DefaultUserNamespace, username))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if skipSweepUserError(err) {
		log.Printf("[WARN] Skipping QuickSight User sweep for %s: %s", region, err.Error())
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Users: %s", err.Error())
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Users for %s: %s", region, err.Error())
	}

	return nil
}

func sweepVPCConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err.Error())
	}

	conn := client.QuickSightClient(ctx)
	awsAccountId := client.AccountID
	sweepResources := make([]sweep.Sweepable, 0)

	input := &quicksight.ListVPCConnectionsInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	out, err := conn.ListVPCConnections(ctx, input)
	for _, v := range out.VPCConnectionSummaries {
		vpcConnectionID := aws.ToString(v.VPCConnectionId)
		sweepResources = append(sweepResources, framework.NewSweepResource(newResourceVPCConnection, client,
			framework.NewAttribute("id", createVPCConnectionID(awsAccountId, vpcConnectionID)),
			framework.NewAttribute("aws_account_id", awsAccountId),
			framework.NewAttribute("vpc_connection_id", vpcConnectionID),
		))
	}

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping QuickSight VPC Connection sweep for %s: %s", region, err.Error())
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight VPC Connections: %s", err.Error())
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight VPC Connections for %s: %s", region, err.Error())
	}

	return nil
}

// skipSweepError adds an additional skippable error code for listing QuickSight resources other than User
func skipSweepError(err error) bool {
	if errs.IsA[*types.UnsupportedUserEditionException](err) {
		return true
	}
	if errs.IsAErrorMessageContains[*types.ResourceNotFoundException](err, "Directory information for account") {
		return true
	}
	if errs.IsAErrorMessageContains[*types.ResourceNotFoundException](err, "Account information for account") {
		return true
	}

	return awsv2.SkipSweepError(err)
}

// skipSweepUserError adds an additional skippable error code for listing QuickSight User resources
func skipSweepUserError(err error) bool {
	if errs.IsAErrorMessageContains[*types.ResourceNotFoundException](err, "not signed up with QuickSight") {
		return true
	}

	if errs.IsAErrorMessageContains[*types.ResourceNotFoundException](err, "Namespace default not found in account") {
		return true
	}

	return awsv2.SkipSweepError(err)
}
