package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCallerIdentity() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCallerIdentityRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCallerIdentityRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).stsconn

	log.Printf("[DEBUG] Reading Caller Identity")
	res, err := client.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		return fmt.Errorf("getting Caller Identity: %w", err)
	}

	log.Printf("[DEBUG] Received Caller Identity: %s", res)

	d.SetId(aws.StringValue(res.Account))
	d.Set("account_id", res.Account)
	d.Set("arn", res.Arn)
	d.Set("user_id", res.UserId)
	d.Set("source_arn", sourceARN(aws.StringValue(res.Arn)))

	return nil
}

// sourceARN returns the same string passed in unless it appears to be an assumed role ARN.
// In that case, it attempts to return the source role ARN associated with an assumed role ARN.
func sourceARN(rawARN string) string {
	result := rawARN

	if strings.Contains(result, ":assumed-role/") && strings.Contains(result, ":sts:") {
		parsedARN, err := arn.Parse(result)

		if err != nil {
			return result
		}

		if parsedARN.Service != "sts" {
			// not an assumed role
			return result
		}

		re := regexp.MustCompile(`^assumed-role/`)
		parsedARN.Resource = re.ReplaceAllString(parsedARN.Resource, "role/")
		parsedARN.Service = "iam"

		if v := strings.LastIndex(parsedARN.Resource, "/"); v > 0 {
			parsedARN.Resource = parsedARN.Resource[0:v]
		}

		result = parsedARN.String()
	}

	return result
}
