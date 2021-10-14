package cognitoidp

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	userPoolDomainDeleteTimeout = 1 * time.Minute
)

// waitUserPoolDomainDeleted waits for an Operation to return Success
func waitUserPoolDomainDeleted(conn *cognitoidentityprovider.CognitoIdentityProvider, domain string) (*cognitoidentityprovider.DescribeUserPoolDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cognitoidentityprovider.DomainStatusTypeUpdating,
			cognitoidentityprovider.DomainStatusTypeDeleting,
		},
		Target:  []string{""},
		Refresh: statusUserPoolDomain(conn, domain),
		Timeout: userPoolDomainDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cognitoidentityprovider.DescribeUserPoolDomainOutput); ok {
		return output, err
	}

	return nil, err
}

func waitUserPoolDomainCreated(conn *cognitoidentityprovider.CognitoIdentityProvider, domain string, timeout time.Duration) (*cognitoidentityprovider.DescribeUserPoolDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cognitoidentityprovider.DomainStatusTypeCreating,
			cognitoidentityprovider.DomainStatusTypeUpdating,
		},
		Target: []string{
			cognitoidentityprovider.DomainStatusTypeActive,
		},
		Refresh: statusUserPoolDomain(conn, domain),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cognitoidentityprovider.DescribeUserPoolDomainOutput); ok {
		return output, err
	}

	return nil, err
}
