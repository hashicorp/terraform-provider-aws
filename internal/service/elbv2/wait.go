package elbv2

import (
	"time"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Default maximum amount of time to wait for a Load Balancer to be created
	loadBalancerCreateTimeout = 10 * time.Minute

	// Default maximum amount of time to wait for a Load Balancer to be updated
	loadBalancerUpdateTimeout = 10 * time.Minute

	// Default maximum amount of time to wait for a Load Balancer to be deleted
	loadBalancerDeleteTimeout = 10 * time.Minute

	// Default maximum amount of time to wait for Tag Propagation for a Load Balancer
	loadBalancerTagPropagationTimeout = 2 * time.Minute

	// Default maximum amount of time to wait for target group to delete
	targetGroupDeleteTimeout = 2 * time.Minute

	// Default maximum amount of time to wait for network interfaces to propagate
	loadBalancerNetworkInterfaceDetachTimeout = 5 * time.Minute

	loadBalancerListenerCreateTimeout = 5 * time.Minute
	loadBalancerListenerReadTimeout   = 2 * time.Minute
	loadBalancerListenerUpdateTimeout = 5 * time.Minute

	propagationTimeout = 2 * time.Minute
)

// waitLoadBalancerActive waits for a Load Balancer to return active
func waitLoadBalancerActive(conn *elbv2.ELBV2, arn string, timeout time.Duration) (*elbv2.LoadBalancer, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{elbv2.LoadBalancerStateEnumProvisioning, elbv2.LoadBalancerStateEnumFailed},
		Target:     []string{elbv2.LoadBalancerStateEnumActive},
		Refresh:    statusLoadBalancerState(conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}
	output, err := stateConf.WaitForState()

	if v, ok := output.(*elbv2.LoadBalancer); ok {
		return v, err
	}
	return nil, err
}
