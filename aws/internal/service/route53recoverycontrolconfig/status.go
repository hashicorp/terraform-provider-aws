package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	Route53RecoveryControlConfigStatusPending         = "PENDING"
	Route53RecoveryControlConfigStatusPendingDeletion = "PENDING_DELETION"
	Route53RecoveryControlConfigStatusDeployed        = "DEPLOYED"
)

func Route53RecoveryControlConfigClusterStatus(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, clusterArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &route53recoverycontrolconfig.DescribeClusterInput{
			ClusterArn: aws.String(clusterArn),
		}

		output, err := conn.DescribeCluster(input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.Cluster.Status), nil
	}
}

func Route53RecoveryControlConfigRoutingControlStatus(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, routingControlArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &route53recoverycontrolconfig.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(routingControlArn),
		}

		output, err := conn.DescribeRoutingControl(input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.RoutingControl.Status), nil
	}
}

func Route53RecoveryControlConfigControlPanelStatus(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, controlPanelArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &route53recoverycontrolconfig.DescribeControlPanelInput{
			ControlPanelArn: aws.String(controlPanelArn),
		}

		output, err := conn.DescribeControlPanel(input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.ControlPanel.Status), nil
	}
}

func Route53RecoveryControlConfigSafetyRuleStatus(conn *route53recoverycontrolconfig.Route53RecoveryControlConfig, safetyRuleArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &route53recoverycontrolconfig.DescribeSafetyRuleInput{
			SafetyRuleArn: aws.String(safetyRuleArn),
		}

		output, err := conn.DescribeSafetyRule(input)

		if err != nil {
			return output, "", err
		}

		if output.AssertionRule != nil {
			return output, aws.StringValue(output.AssertionRule.Status), nil
		}

		if output.GatingRule != nil {
			return output, aws.StringValue(output.GatingRule.Status), nil
		}

		return output, "", nil
	}
}
