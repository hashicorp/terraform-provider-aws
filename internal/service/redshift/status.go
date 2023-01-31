package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusClusterAvailability(ctx context.Context, conn *redshift.Redshift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ClusterAvailabilityStatus), nil
	}
}

func statusClusterAvailabilityZoneRelocation(ctx context.Context, conn *redshift.Redshift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.AvailabilityZoneRelocationStatus), nil
	}
}

func statusCluster(ctx context.Context, conn *redshift.Redshift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ClusterStatus), nil
	}
}

func statusClusterAqua(ctx context.Context, conn *redshift.Redshift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.AquaConfiguration.AquaStatus), nil
	}
}

func statusScheduleAssociation(ctx context.Context, conn *redshift.Redshift, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, output, err := FindScheduleAssociationById(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ScheduleAssociationState), nil
	}
}

func statusEndpointAccess(ctx context.Context, conn *redshift.Redshift, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEndpointAccessByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.EndpointStatus), nil
	}
}
