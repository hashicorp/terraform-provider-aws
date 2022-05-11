package route53recoverycontrolconfig

import (
	"time"

	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	route53RecoveryControlConfigTimeout    = 60 * time.Second
	route53RecoveryControlConfigMinTimeout = 5 * time.Second
)

func waitClusterCreated(conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusCluster(conn, clusterArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusCluster(conn, clusterArn),
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

func waitRoutingControlCreated(conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusRoutingControl(conn, routingControlArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoutingControlDeleted(conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusRoutingControl(conn, routingControlArn),
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

func waitControlPanelCreated(conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusControlPanel(conn, controlPanelArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func waitControlPanelDeleted(conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusControlPanel(conn, controlPanelArn),
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

func waitSafetyRuleCreated(conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusSafetyRule(conn, safetyRuleArn),
		Timeout:    route53RecoveryControlConfigTimeout,
		MinTimeout: route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*r53rcc.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func waitSafetyRuleDeleted(conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusSafetyRule(conn, safetyRuleArn),
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
