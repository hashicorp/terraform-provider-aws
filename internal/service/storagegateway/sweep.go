//go:build sweep
// +build sweep

package storagegateway

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_storagegateway_gateway", &resource.Sweeper{
		Name: "aws_storagegateway_gateway",
		F:    sweepGateways,
	})
}

func sweepGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).StorageGatewayConn

	err = conn.ListGatewaysPages(&storagegateway.ListGatewaysInput{}, func(page *storagegateway.ListGatewaysOutput, lastPage bool) bool {
		if len(page.Gateways) == 0 {
			log.Print("[DEBUG] No Storage Gateway Gateways to sweep")
			return true
		}

		for _, gateway := range page.Gateways {
			name := aws.StringValue(gateway.GatewayName)

			log.Printf("[INFO] Deleting Storage Gateway Gateway: %s", name)
			input := &storagegateway.DeleteGatewayInput{
				GatewayARN: gateway.GatewayARN,
			}

			_, err := conn.DeleteGateway(input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, storagegateway.ErrorCodeGatewayNotFound, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Storage Gateway Gateway (%s): %s", name, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Storage Gateway Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Storage Gateway Gateways: %w", err)
	}
	return nil
}
