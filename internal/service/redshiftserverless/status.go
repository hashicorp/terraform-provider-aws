package redshiftserverless

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusNamespace(conn *redshiftserverless.RedshiftServerless, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNamespaceByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusWorkgroup(conn *redshiftserverless.RedshiftServerless, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindWorkgroupByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
