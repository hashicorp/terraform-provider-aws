package cognitoidp

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	userPoolDomainDeleteTimeout = 1 * time.Minute
)

// waitUserPoolDomainDeleted waits for an Operation to return Success
func waitUserPoolDomainDeleted(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, domain string) (*cognitoidentityprovider.DescribeUserPoolDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cognitoidentityprovider.DomainStatusTypeUpdating,
			cognitoidentityprovider.DomainStatusTypeDeleting,
		},
		Target:  []string{""},
		Refresh: statusUserPoolDomain(ctx, conn, domain),
		Timeout: userPoolDomainDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cognitoidentityprovider.DescribeUserPoolDomainOutput); ok {
		return output, err
	}

	return nil, err
}

func waitUserPoolDomainCreated(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, domain string, timeout time.Duration) (*cognitoidentityprovider.DescribeUserPoolDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cognitoidentityprovider.DomainStatusTypeCreating,
			cognitoidentityprovider.DomainStatusTypeUpdating,
		},
		Target: []string{
			cognitoidentityprovider.DomainStatusTypeActive,
		},
		Refresh: statusUserPoolDomain(ctx, conn, domain),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cognitoidentityprovider.DescribeUserPoolDomainOutput); ok {
		return output, err
	}

	return nil, err
}
