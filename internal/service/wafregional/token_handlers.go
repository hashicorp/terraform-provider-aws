package wafregional

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

type WafRegionalRetryer struct {
	Connection *wafregional.WAFRegional
	Region     string
}

type withRegionalTokenFunc func(token *string) (interface{}, error)

func (t *WafRegionalRetryer) RetryWithToken(f withRegionalTokenFunc) (interface{}, error) {
	conns.GlobalMutexKV.Lock(t.Region)
	defer conns.GlobalMutexKV.Unlock(t.Region)

	var out interface{}
	var tokenOut *waf.GetChangeTokenOutput
	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error

		tokenOut, err = t.Connection.GetChangeToken(&waf.GetChangeTokenInput{})
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Failed to acquire change token: %w", err))
		}

		out, err = f(tokenOut.ChangeToken)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, waf.ErrCodeStaleDataException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		tokenOut, err = t.Connection.GetChangeToken(&waf.GetChangeTokenInput{})

		if err != nil {
			return nil, fmt.Errorf("error getting WAF Regional change token: %w", err)
		}

		out, err = f(tokenOut.ChangeToken)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func newWafRegionalRetryer(conn *wafregional.WAFRegional, region string) *WafRegionalRetryer {
	return &WafRegionalRetryer{Connection: conn, Region: region}
}
