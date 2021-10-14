package waiter

import (
	"time"

	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	route53RecoveryControlConfigTimeout    = 60 * time.Second
	route53RecoveryControlConfigMinTimeout = 5 * time.Second
)

func waitRoute53RecoveryControlConfigClusterCreated(conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusRoute53RecoveryControlConfigCluster(conn, clusterArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoute53RecoveryControlConfigClusterDeleted(conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusRoute53RecoveryControlConfigCluster(conn, clusterArn),
		Timeout:        route53RecoveryControlConfigTimeout,
		Delay:          route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoute53RecoveryControlConfigRoutingControlCreated(conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusRoute53RecoveryControlConfigRoutingControl(conn, routingControlArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoute53RecoveryControlConfigRoutingControlDeleted(conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusRoute53RecoveryControlConfigRoutingControl(conn, routingControlArn),
		Timeout:        route53RecoveryControlConfigTimeout,
		Delay:          route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoute53RecoveryControlConfigControlPanelCreated(conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusRoute53RecoveryControlConfigControlPanel(conn, controlPanelArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoute53RecoveryControlConfigControlPanelDeleted(conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusRoute53RecoveryControlConfigControlPanel(conn, controlPanelArn),
		Timeout:        route53RecoveryControlConfigTimeout,
		Delay:          route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoute53RecoveryControlConfigSafetyRuleCreated(conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusRoute53RecoveryControlConfigSafetyRule(conn, safetyRuleArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoute53RecoveryControlConfigSafetyRuleDeleted(conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusRoute53RecoveryControlConfigSafetyRule(conn, safetyRuleArn),
		Timeout:        route53RecoveryControlConfigTimeout,
		Delay:          route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}
