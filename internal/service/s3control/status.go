package s3control

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusMultiRegionAccessPointRequest(conn *s3control.S3Control, accountID string, requestTokenARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findMultiRegionAccessPointOperationByAccountIDAndTokenARN(conn, accountID, requestTokenARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.RequestStatus), nil
	}
}
