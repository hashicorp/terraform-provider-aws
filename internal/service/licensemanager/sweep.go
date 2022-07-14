//go:build sweep
// +build sweep

package licensemanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_licensemanager_license_configuration", &resource.Sweeper{
		Name: "aws_licensemanager_license_configuration",
		F:    sweepLicenseConfigurations,
	})
}

func sweepLicenseConfigurations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LicenseManagerConn

	resp, err := conn.ListLicenseConfigurations(&licensemanager.ListLicenseConfigurationsInput{})

	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping License Manager License Configuration sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving License Manager license configurations: %s", err)
	}

	if len(resp.LicenseConfigurations) == 0 {
		log.Print("[DEBUG] No License Manager license configurations to sweep")
		return nil
	}

	for _, lc := range resp.LicenseConfigurations {
		id := aws.StringValue(lc.LicenseConfigurationArn)

		log.Printf("[INFO] Deleting License Manager license configuration: %s", id)

		opts := &licensemanager.DeleteLicenseConfigurationInput{
			LicenseConfigurationArn: aws.String(id),
		}

		_, err := conn.DeleteLicenseConfiguration(opts)

		if err != nil {
			log.Printf("[ERROR] Error deleting License Manager license configuration (%s): %s", id, err)
		}
	}

	return nil
}
