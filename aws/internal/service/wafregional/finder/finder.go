package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func RegexMatchSetByID(conn *wafregional.WAFRegional, id string) (*waf.RegexMatchSet, error) {
	result, err := conn.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
		RegexMatchSetId: aws.String(id),
	})

	return result.RegexMatchSet, err
}
