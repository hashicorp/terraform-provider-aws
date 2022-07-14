package appsync

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAPICacheByID(conn *appsync.AppSync, id string) (*appsync.ApiCache, error) {
	input := &appsync.GetApiCacheInput{
		ApiId: aws.String(id),
	}
	out, err := conn.GetApiCache(input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.ApiCache, nil
}

func FindDomainNameByID(conn *appsync.AppSync, id string) (*appsync.DomainNameConfig, error) {
	input := &appsync.GetDomainNameInput{
		DomainName: aws.String(id),
	}
	out, err := conn.GetDomainName(input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.DomainNameConfig, nil
}

func FindDomainNameAPIAssociationByID(conn *appsync.AppSync, id string) (*appsync.ApiAssociation, error) {
	input := &appsync.GetApiAssociationInput{
		DomainName: aws.String(id),
	}
	out, err := conn.GetApiAssociation(input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.ApiAssociation, nil
}
