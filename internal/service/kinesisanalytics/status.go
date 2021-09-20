package kinesisanalytics

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/kinesisanalytics/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// statusApplication fetches the ApplicationDetail and its Status
func statusApplication(conn *kinesisanalytics.KinesisAnalytics, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		applicationDetail, err := finder.FindApplicationDetailByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return applicationDetail, aws.StringValue(applicationDetail.ApplicationStatus), nil
	}
}
