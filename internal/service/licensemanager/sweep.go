// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_licensemanager_license_configuration", &resource.Sweeper{
		Name: "aws_licensemanager_license_configuration",
		F:    sweepLicenseConfigurations,
	})
}

func sweepLicenseConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.LicenseManagerClient(ctx)
	input := &licensemanager.ListLicenseConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listLicenseConfigurationsPages(ctx, conn, input, func(page *licensemanager.ListLicenseConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LicenseConfigurations {
			r := resourceLicenseConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.LicenseConfigurationArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping License Manager License Configuration sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing License Manager License Configurations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping License Manager License Configurations (%s): %w", region, err)
	}

	return nil
}
