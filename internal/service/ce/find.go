package ce

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAnomalyMonitorByARN(ctx context.Context, conn *costexplorer.CostExplorer, arn string) (*costexplorer.AnomalyMonitor, error) {
	in := &costexplorer.GetAnomalyMonitorsInput{
		MonitorArnList: aws.StringSlice([]string{arn}),
		MaxResults:     aws.Int64(1),
	}

	out, err := conn.GetAnomalyMonitorsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeUnknownMonitorException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.AnomalyMonitors) == 0 || out.AnomalyMonitors[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AnomalyMonitors[0], nil
}

func FindAnomalySubscriptionByARN(ctx context.Context, conn *costexplorer.CostExplorer, arn string) (*costexplorer.AnomalySubscription, error) {
	in := &costexplorer.GetAnomalySubscriptionsInput{
		SubscriptionArnList: aws.StringSlice([]string{arn}),
		MaxResults:          aws.Int64(1),
	}

	out, err := conn.GetAnomalySubscriptionsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeUnknownMonitorException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.AnomalySubscriptions) == 0 || out.AnomalySubscriptions[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AnomalySubscriptions[0], nil
}

func FindCostAllocationTagByKey(ctx context.Context, conn *costexplorer.CostExplorer, key string) (*costexplorer.CostAllocationTag, error) {
	in := &costexplorer.ListCostAllocationTagsInput{
		TagKeys:    aws.StringSlice([]string{key}),
		MaxResults: aws.Int64(1),
	}

	out, err := conn.ListCostAllocationTagsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeUnknownMonitorException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.CostAllocationTags) == 0 || out.CostAllocationTags[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.CostAllocationTags[0], nil
}
