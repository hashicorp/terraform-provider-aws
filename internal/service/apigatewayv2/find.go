package apigatewayv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// FindAPIByID returns the API corresponding to the specified ID.
// Returns NotFoundError if no API is found.
func FindAPIByID(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, apiID string) (*apigatewayv2.GetApiOutput, error) {
	input := &apigatewayv2.GetApiInput{
		ApiId: aws.String(apiID),
	}

	return FindAPI(ctx, conn, input)
}

// FindAPI returns the API corresponding to the specified input.
// Returns NotFoundError if no API is found.
func FindAPI(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApiInput) (*apigatewayv2.GetApiOutput, error) {
	output, err := conn.GetApiWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

// FindAPIs returns the APIs corresponding to the specified input.
// Returns an empty slice if no APIs are found.
func FindAPIs(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApisInput) ([]*apigatewayv2.Api, error) {
	var apis []*apigatewayv2.Api

	err := getAPIsPages(ctx, conn, input, func(page *apigatewayv2.GetApisOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			apis = append(apis, item)
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return apis, nil
}

func FindDomainNameByName(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, name string) (*apigatewayv2.GetDomainNameOutput, error) {
	input := &apigatewayv2.GetDomainNameInput{
		DomainName: aws.String(name),
	}

	return FindDomainName(ctx, conn, input)
}

func FindDomainName(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetDomainNameInput) (*apigatewayv2.GetDomainNameOutput, error) {
	output, err := conn.GetDomainNameWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if output == nil || len(output.DomainNameConfigurations) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
