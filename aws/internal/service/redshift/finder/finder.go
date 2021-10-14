package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ClusterByID(conn *redshift.Redshift, id string) (*redshift.Cluster, error) {
	input := &redshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(id),
	}

	output, err := conn.DescribeClusters(input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Clusters) == 0 || output.Clusters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Clusters); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Clusters[0], nil
}

func ScheduledActionByName(conn *redshift.Redshift, name string) (*redshift.ScheduledAction, error) {
	input := &redshift.DescribeScheduledActionsInput{
		ScheduledActionName: aws.String(name),
	}

	output, err := conn.DescribeScheduledActions(input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeScheduledActionNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ScheduledActions) == 0 || output.ScheduledActions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ScheduledActions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ScheduledActions[0], nil
}
