package redshiftserverless

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindNamespaceByName(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Namespace, error) {
	input := &redshiftserverless.GetNamespaceInput{
		NamespaceName: aws.String(name),
	}

	output, err := conn.GetNamespace(input)

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

func FindWorkgroupByName(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Workgroup, error) {
	input := &redshiftserverless.GetWorkgroupInput{
		WorkgroupName: aws.String(name),
	}

	output, err := conn.GetWorkgroup(input)

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

func FindEndpointAccessByName(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.EndpointAccess, error) {
	input := &redshiftserverless.GetEndpointAccessInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.GetEndpointAccess(input)

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
