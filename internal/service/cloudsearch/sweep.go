//go:build sweep
// +build sweep

package cloudsearch

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudsearch_domain", &resource.Sweeper{
		Name: "aws_cloudsearch_domain",
		F: func(region string) error {
			client, err := sweep.SharedRegionalSweepClient(region)
			if err != nil {
				return fmt.Errorf("error getting client: %w", err)
			}
			conn := client.(*conns.AWSClient).CloudSearchConn

			domains, err := conn.DescribeDomains(&cloudsearch.DescribeDomainsInput{})
			if err != nil {
				return fmt.Errorf("error describing CloudSearch domains: %s", err)
			}

			for _, domain := range domains.DomainStatusList {
				if !strings.HasPrefix(*domain.DomainName, "tf-acc-") {
					continue
				}
				_, err := conn.DeleteDomain(&cloudsearch.DeleteDomainInput{
					DomainName: domain.DomainName,
				})
				if err != nil {
					return fmt.Errorf("error deleting CloudSearch domain: %s", err)
				}
			}
			return nil
		},
	})
}
