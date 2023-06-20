package vpclattice

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// findServiceByAttributes finds a first service that matches the given service name.
func findServiceByAttributes(ctx context.Context, conn *vpclattice.Client, serviceName string) (*vpclattice.GetServiceOutput, error) {
	var serviceSummary *types.ServiceSummary

	input := &vpclattice.ListServicesInput{}
	paginator := vpclattice.NewListServicesPaginator(conn, input, func(options *vpclattice.ListServicesPaginatorOptions) {
		options.Limit = 100
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			if serviceName == *v.Name {
				serviceSummary = &v
				break
			}
		}
	}

	if serviceSummary == nil {
		return nil, &retry.NotFoundError{
			Message: "No matching EC2 VPC Lattice Service found",
		}
	}

	service, err := conn.GetService(ctx, &vpclattice.GetServiceInput{
		ServiceIdentifier: serviceSummary.Id,
	})

	if err != nil {
		return nil, err
	}

	return service, nil
}
