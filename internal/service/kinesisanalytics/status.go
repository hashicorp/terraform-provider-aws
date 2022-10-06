package kinesisanalytics

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusApplication fetches the ApplicationDetail and its Status
func statusApplication(conn *kinesisanalytics.KinesisAnalytics, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		applicationDetail, err := FindApplicationDetailByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return applicationDetail, aws.StringValue(applicationDetail.ApplicationStatus), nil
	}
}
