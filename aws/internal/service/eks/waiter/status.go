package waiter

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ResourceStatusFailed  = "Failed"
	ResourceStatusUnknown = "Unknown"
	ResourceStatusDeleted = "Deleted"
)

func EksAddonCreatedStatus(ctx context.Context, conn *eks.EKS, addonName, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			ClusterName: aws.String(clusterName),
			AddonName:   aws.String(addonName),
		})
		if err != nil {
			return output, ResourceStatusFailed, err
		}
		if output == nil || output.Addon == nil {
			return nil, ResourceStatusUnknown, fmt.Errorf("EKS Cluster (%s) Addon (%s) missing", clusterName, addonName)
		}
		return output.Addon, aws.StringValue(output.Addon.Status), nil
	}
}

func EksAddonDeletedStatus(ctx context.Context, conn *eks.EKS, addonName, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			ClusterName: aws.String(clusterName),
			AddonName:   aws.String(addonName),
		})
		if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
			return output, ResourceStatusDeleted, nil
		}
		if err != nil {
			return output, ResourceStatusFailed, err
		}
		if output == nil || output.Addon == nil {
			return output, ResourceStatusUnknown, fmt.Errorf("EKS Cluster (%s) Addon (%s) missing", clusterName, addonName)
		}
		return output.Addon, aws.StringValue(output.Addon.Status), nil
	}
}

func EksAddonUpdateStatus(ctx context.Context, conn *eks.EKS, clusterName, addonName, updateID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeUpdateWithContext(ctx, &eks.DescribeUpdateInput{
			Name:      aws.String(clusterName),
			AddonName: aws.String(addonName),
			UpdateId:  aws.String(updateID),
		})
		if err != nil {
			return output, ResourceStatusFailed, err
		}
		if output == nil || output.Update == nil {
			return nil, ResourceStatusUnknown, fmt.Errorf("EKS Cluster (%s) Addon (%s) update (%s) missing", clusterName, addonName, updateID)
		}
		return output.Update, aws.StringValue(output.Update.Status), nil
	}
}
