package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/lister"
)

// ApiByID returns the API corresponding to the specified ID.
func ApiByID(conn *apigatewayv2.ApiGatewayV2, apiID string) (*apigatewayv2.GetApiOutput, error) {
	input := &apigatewayv2.GetApiInput{
		ApiId: aws.String(apiID),
	}

	output, err := conn.GetApi(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// Apis returns the APIs corresponding to the specified input.
// Returns an empty slice if no APIs are found.
func Apis(conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApisInput) ([]*apigatewayv2.Api, error) {
	var apis []*apigatewayv2.Api

	err := lister.GetApisPages(conn, input, func(page *apigatewayv2.GetApisOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			apis = append(apis, item)
		}

		return !isLast
	})

	if err != nil {
		return nil, err
	}

	return apis, nil
}
