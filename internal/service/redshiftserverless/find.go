package redshiftserverless

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindNamespaceByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Namespace, error) {
	input := &redshiftserverless.GetNamespaceInput{
		NamespaceName: aws.String(name),
	}

	output, err := conn.GetNamespaceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Namespace, nil
}

func FindWorkgroupByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Workgroup, error) {
	input := &redshiftserverless.GetWorkgroupInput{
		WorkgroupName: aws.String(name),
	}

	output, err := conn.GetWorkgroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workgroup, nil
}

func FindEndpointAccessByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.EndpointAccess, error) {
	input := &redshiftserverless.GetEndpointAccessInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.GetEndpointAccessWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Endpoint, nil
}

func FindUsageLimitByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, id string) (*redshiftserverless.UsageLimit, error) {
	input := &redshiftserverless.GetUsageLimitInput{
		UsageLimitId: aws.String(id),
	}

	output, err := conn.GetUsageLimitWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, redshiftserverless.ErrCodeValidationException, "does not exist") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.UsageLimit, nil
}

func FindSnapshotByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Snapshot, error) {
	input := &redshiftserverless.GetSnapshotInput{
		SnapshotName: aws.String(name),
	}

	output, err := conn.GetSnapshotWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, redshiftserverless.ErrCodeResourceNotFoundException, "snapshot") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Snapshot, nil
}

func FindResourcePolicyByARN(ctx context.Context, conn *redshiftserverless.RedshiftServerless, arn string) (*redshiftserverless.ResourcePolicy, error) {
	input := &redshiftserverless.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicyWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, redshiftserverless.ErrCodeResourceNotFoundException, "does not exist") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourcePolicy, nil
}
