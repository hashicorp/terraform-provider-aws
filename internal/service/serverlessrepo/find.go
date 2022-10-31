package serverlessrepo

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	serverlessrepo "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func findApplication(conn *serverlessrepo.ServerlessApplicationRepository, applicationID, version string) (*serverlessrepo.GetApplicationOutput, error) {
	input := &serverlessrepo.GetApplicationInput{
		ApplicationId: aws.String(applicationID),
	}
	if version != "" {
		input.SemanticVersion = aws.String(version)
	}

	log.Printf("[DEBUG] Getting Serverless findApplication Repository Application: %s", input)
	resp, err := conn.GetApplication(input)
	if tfawserr.ErrCodeEquals(err, serverlessrepo.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:    err,
			LastRequest:  input,
			LastResponse: resp,
		}
	}
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, &resource.NotFoundError{
			LastRequest:  input,
			LastResponse: resp,
			Message:      "returned empty response",
		}
	}

	return resp, nil
}
