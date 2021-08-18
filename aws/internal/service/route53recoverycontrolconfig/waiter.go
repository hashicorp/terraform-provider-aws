package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	Route53RecoveryControlConfigTimeout    = 60 * time.Second
	Route53RecoveryControlConfigMinTimeout = 5 * time.Second
)

func Route53RecoveryControlConfigClusterCreated(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, clusterArn string) (*route53recoverycontrolconfig.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{Route53RecoveryControlConfigStatusPending},
		Target:     []string{Route53RecoveryControlConfigStatusDeployed},
		Refresh:    Route53RecoveryControlConfigClusterStatus(conn, clusterArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigClusterDeleted(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, clusterArn string) (*route53recoverycontrolconfig.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{Route53RecoveryControlConfigStatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigClusterStatus(conn, clusterArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigRoutingControlCreated(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, routingControlArn string) (*route53recoverycontrolconfig.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{Route53RecoveryControlConfigStatusPending},
		Target:     []string{Route53RecoveryControlConfigStatusDeployed},
		Refresh:    Route53RecoveryControlConfigRoutingControlStatus(conn, routingControlArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigRoutingControlDeleted(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, routingControlArn string) (*route53recoverycontrolconfig.DescribeRoutingControlOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{Route53RecoveryControlConfigStatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigRoutingControlStatus(conn, routingControlArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigControlPanelCreated(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, controlPanelArn string) (*route53recoverycontrolconfig.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{Route53RecoveryControlConfigStatusPending},
		Target:     []string{Route53RecoveryControlConfigStatusDeployed},
		Refresh:    Route53RecoveryControlConfigControlPanelStatus(conn, controlPanelArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigControlPanelDeleted(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, controlPanelArn string) (*route53recoverycontrolconfig.DescribeControlPanelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{Route53RecoveryControlConfigStatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigControlPanelStatus(conn, controlPanelArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigSafetyRuleCreated(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, safetyRuleArn string) (*route53recoverycontrolconfig.DescribeSafetyRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{Route53RecoveryControlConfigStatusPending},
		Target:     []string{Route53RecoveryControlConfigStatusDeployed},
		Refresh:    Route53RecoveryControlConfigSafetyRuleStatus(conn, safetyRuleArn),
		Timeout:    Route53RecoveryControlConfigTimeout,
		MinTimeout: Route53RecoveryControlConfigMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func Route53RecoveryControlConfigSafetyRuleDeleted(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, safetyRuleArn string) (*route53recoverycontrolconfig.DescribeSafetyRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{Route53RecoveryControlConfigStatusPendingDeletion},
		Target:         []string{},
		Refresh:        Route53RecoveryControlConfigSafetyRuleStatus(conn, safetyRuleArn),
		Timeout:        Route53RecoveryControlConfigTimeout,
		Delay:          Route53RecoveryControlConfigMinTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53recoverycontrolconfig.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}
