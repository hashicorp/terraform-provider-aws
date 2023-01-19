//go:build sweep
// +build sweep

package elb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_elb", &resource.Sweeper{
		Name: "aws_elb",
		F:    sweepLoadBalancers,
	})
}

func sweepLoadBalancers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ELBConn()

	err = conn.DescribeLoadBalancersPagesWithContext(ctx, &elb.DescribeLoadBalancersInput{}, func(out *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
		if len(out.LoadBalancerDescriptions) == 0 {
			log.Println("[INFO] No ELBs found for sweeping")
			return false
		}

		for _, lb := range out.LoadBalancerDescriptions {
			log.Printf("[INFO] Deleting ELB: %s", *lb.LoadBalancerName)

			_, err := conn.DeleteLoadBalancerWithContext(ctx, &elb.DeleteLoadBalancerInput{
				LoadBalancerName: lb.LoadBalancerName,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete ELB %s: %s", *lb.LoadBalancerName, err)
				continue
			}
			err = CleanupNetworkInterfaces(ctx, client.(*conns.AWSClient).EC2Conn(), *lb.LoadBalancerName)
			if err != nil {
				log.Printf("[WARN] Failed to cleanup ENIs for ELB %q: %s", *lb.LoadBalancerName, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ELB sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving ELBs: %s", err)
	}
	return nil
}
