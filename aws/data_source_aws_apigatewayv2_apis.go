package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsApiGatewayV2Apis() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAwsApiGatewayV2ApisRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsAwsApiGatewayV2ApisRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	tagsToMatch := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	var ids []*string

	err := ListAllApiGatewayV2Apis(conn, func(page *apigatewayv2.GetApisOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, api := range page.Items {
			if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(api.Name) {
				continue
			}

			if v, ok := d.GetOk("protocol_type"); ok && v.(string) != aws.StringValue(api.ProtocolType) {
				continue
			}

			if len(tagsToMatch) > 0 && !keyvaluetags.Apigatewayv2KeyValueTags(api.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).ContainsAll(tagsToMatch) {
				continue
			}

			ids = append(ids, api.ApiId)
		}

		return !isLast
	})

	if err != nil {
		return fmt.Errorf("error listing API Gateway v2 APIs: %w", err)
	}

	d.SetId(resource.UniqueId())

	if err := d.Set("ids", flattenStringSet(ids)); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}

//
// These can be moved to a per-service package like "aws/internal/service/apigatewayv2/lister" in the future.
//

func ListAllApiGatewayV2Apis(conn *apigatewayv2.ApiGatewayV2, fn func(*apigatewayv2.GetApisOutput, bool) bool) error {
	return ListApiGatewayV2ApisPages(conn, &apigatewayv2.GetApisInput{}, fn)
}

func ListApiGatewayV2ApisPages(conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApisInput, fn func(*apigatewayv2.GetApisOutput, bool) bool) error {
	for {
		output, err := conn.GetApis(input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
