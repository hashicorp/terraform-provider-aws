package batch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func statusComputeEnvironment(conn *batch.Batch, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		computeEnvironmentDetail, err := tfbatch.FindComputeEnvironmentDetailByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return computeEnvironmentDetail, aws.StringValue(computeEnvironmentDetail.Status), nil
	}
}
