package opensearch

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDomainByName(ctx context.Context, conn *opensearchservice.OpenSearchService, name string) (*opensearchservice.DomainStatus, error) {
	input := &opensearchservice.DescribeDomainInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeDomainWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainStatus, nil
}
