package elasticsearch

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDomainByName(conn *elasticsearch.ElasticsearchService, name string) (*elasticsearch.ElasticsearchDomainStatus, error) {
	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeElasticsearchDomain(input)

	if tfawserr.ErrCodeEquals(err, elasticsearch.ErrCodeResourceNotFoundException) {
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

func FindVPCEndpointByID(ctx context.Context, conn *elasticsearch.ElasticsearchService, id string) (*elasticsearch.VpcEndpoint, error) {
	output, err := conn.DescribeVpcEndpointsWithContext(ctx, &elasticsearch.DescribeVpcEndpointsInput{
		VpcEndpointIds: []*string{&id},
	})
	if err != nil {
		return nil, err
	}

	countVPCEndpoints := len(output.VpcEndpoints)
	if countVPCEndpoints == 0 {
		return nil, fmt.Errorf("got more than one VPCEndpoint for id ( %s )", id)
	}
	if countVPCEndpoints > 1 {
		return output.VpcEndpoints[0], fmt.Errorf("got %d instead of one VPCEndpoint for id ( %s )", countVPCEndpoints, id)
	}
	return output.VpcEndpoints[0], nil
}
