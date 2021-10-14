package waiter

import (
	"time"

	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	Route53RecoveryControlConfigTimeout    = 60 * time.Second
	Route53RecoveryControlConfigMinTimeout = 5 * time.Second
)

func Route53RecoveryControlConfigClusterCreated(conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    Route53RecoveryControlConfigClusterStatus(conn, clusterArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigClusterDeleted(conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigClusterStatus(conn, clusterArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigRoutingControlCreated(conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    Route53RecoveryControlConfigRoutingControlStatus(conn, routingControlArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigRoutingControlDeleted(conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigRoutingControlStatus(conn, routingControlArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigControlPanelCreated(conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    Route53RecoveryControlConfigControlPanelStatus(conn, controlPanelArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigControlPanelDeleted(conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigControlPanelStatus(conn, controlPanelArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigSafetyRuleCreated(conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    Route53RecoveryControlConfigSafetyRuleStatus(conn, safetyRuleArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigSafetyRuleDeleted(conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigSafetyRuleStatus(conn, safetyRuleArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}
