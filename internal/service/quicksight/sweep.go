// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package quicksight

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	resource.AddTestSweepers("aws_quicksight_template", &resource.Sweeper{
		Name: "aws_quicksight_template",
		F:    sweepTemplates,
	})
	resource.AddTestSweepers("aws_quicksight_user", &resource.Sweeper{
		Name: "aws_quicksight_user",
		F:    sweepUsers,
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

	conn := client.QuickSightConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListDashboardsInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	err = conn.ListDashboardsPagesWithContext(ctx, input, func(page *quicksight.ListDashboardsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dashboard := range page.DashboardSummaryList {
			if dashboard == nil {
				continue
			}

			r := ResourceDashboard()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.StringValue(dashboard.DashboardId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping QuickSight Dashboard sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Dashboards: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Dashboards for %s: %w", region, err)
	}

	return nil
}

func sweepDataSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.QuickSightConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListDataSetsInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	err = conn.ListDataSetsPagesWithContext(ctx, input, func(page *quicksight.ListDataSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ds := range page.DataSetSummaries {
			if ds == nil {
				continue
			}

			r := ResourceDataSet()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.StringValue(ds.DataSetId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping QuickSight Data Set sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Data Sets: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Data Sets for %s: %w", region, err)
	}

	return nil
}

func sweepDataSources(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.QuickSightConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListDataSourcesInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	err = conn.ListDataSourcesPagesWithContext(ctx, input, func(page *quicksight.ListDataSourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ds := range page.DataSources {
			if ds == nil {
				continue
			}

			r := ResourceDataSource()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s/%s", awsAccountId, aws.StringValue(ds.DataSourceId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping QuickSight Data Source sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Data Sources: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Data Sources for %s: %w", region, err)
	}

	return nil
}

func sweepFolders(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.QuickSightConn(ctx)
	awsAccountId := client.AccountID
	sweepResources := make([]sweep.Sweepable, 0)

	input := &quicksight.ListFoldersInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	out, err := conn.ListFoldersWithContext(ctx, input)
	for _, folder := range out.FolderSummaryList {
		if folder.FolderId == nil {
			continue
		}

		r := ResourceFolder()
		d := r.Data(nil)
		d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.StringValue(folder.FolderId)))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping QuickSight Folder sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Folders: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Folders for %s: %w", region, err)
	}

	return nil

}

func sweepTemplates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.QuickSightConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	awsAccountId := client.AccountID

	input := &quicksight.ListTemplatesInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	err = conn.ListTemplatesPagesWithContext(ctx, input, func(page *quicksight.ListTemplatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, tmpl := range page.TemplateSummaryList {
			if tmpl == nil {
				continue
			}

			r := ResourceTemplate()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", awsAccountId, aws.StringValue(tmpl.TemplateId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping QuickSight Template sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Templates: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Templates for %s: %w", region, err)
	}

	return nil
}

func sweepUsers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.QuickSightConn(ctx)
	awsAccountId := client.AccountID
	sweepResources := make([]sweep.Sweepable, 0)

	input := &quicksight.ListUsersInput{
		AwsAccountId: aws.String(awsAccountId),
		Namespace:    aws.String(DefaultUserNamespace),
	}

	out, err := conn.ListUsersWithContext(ctx, input)
	for _, user := range out.UserList {
		username := aws.StringValue(user.UserName)
		if !strings.HasPrefix(username, acctestResourcePrefix) {
			continue
		}

		r := ResourceUser()
		d := r.Data(nil)
		d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountId, DefaultUserNamespace, username))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if skipSweepUserError(err) {
		log.Printf("[WARN] Skipping QuickSight User sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing QuickSight Users: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping QuickSight Users for %s: %w", region, err)
	}

	return nil
}

// skipSweepError adds an additional skippable error code for listing QuickSight resources other than User
func skipSweepError(err error) bool {
	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeUnsupportedUserEditionException) {
		return true
	}

	return sweep.SkipSweepError(err)
}

// skipSweepUserError adds an additional skippable error code for listing QuickSight User resources
func skipSweepUserError(err error) bool {
	if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "not signed up with QuickSight") {
		return true
	}

	return sweep.SkipSweepError(err)

}
