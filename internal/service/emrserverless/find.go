package emrserverless

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindApplicationByID(conn *emrserverless.EMRServerless, id string) (*emrserverless.Application, error) {
	input := &emrserverless.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApplication(input)

	if tfawserr.ErrCodeEquals(err, emrserverless.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Application == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if aws.StringValue(output.Application.State) == emrserverless.ApplicationStateTerminated {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Application, nil
}
