package appsync

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusAPICache(conn *appsync.AppSync, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAPICacheByID(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusDomainNameAPIAssociation(conn *appsync.AppSync, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDomainNameAPIAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.AssociationStatus), nil
	}
}
