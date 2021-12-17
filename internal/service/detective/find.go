package detective

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindDetectiveGraphByArn(conn *detective.Detective, ctx context.Context, arn string) (*detective.Graph, error) {
	input := &detective.ListGraphsInput{}
	var result *detective.Graph

	err := conn.ListGraphsPagesWithContext(ctx, input, func(page *detective.ListGraphsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, graph := range page.GraphList {
			if graph == nil {
				continue
			}

			if aws.StringValue(graph.Arn) == arn {
				result = graph
				return false
			}
		}
		return !lastPage
	})
	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     fmt.Sprintf("No detective graph with arn %q", arn),
			LastRequest: input,
		}
	}

	return result, nil
}
