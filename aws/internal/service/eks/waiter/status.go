package waiter

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func ClusterStatus(conn *eks.EKS, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ClusterByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func FargateProfileStatus(conn *eks.EKS, clusterName, fargateProfileName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FargateProfileByClusterNameAndFargateProfileName(conn, clusterName, fargateProfileName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func UpdateStatus(conn *eks.EKS, name, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.UpdateByNameAndID(conn, name, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func EksAddonStatus(ctx context.Context, conn *eks.EKS, addonName, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			ClusterName: aws.String(clusterName),
			AddonName:   aws.String(addonName),
		})
		if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}
		if err != nil {
			return output, "", err
		}
		if output == nil || output.Addon == nil {
			return nil, "", fmt.Errorf("EKS Cluster (%s) add-on (%s) missing", clusterName, addonName)
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
			return output, "", err
		}
		if output == nil || output.Update == nil {
			return nil, "", fmt.Errorf("EKS Cluster (%s) add-on (%s) update (%s) missing", clusterName, addonName, updateID)
		}
		return output.Update, aws.StringValue(output.Update.Status), nil
	}
}
