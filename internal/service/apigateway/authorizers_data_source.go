package apigateway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_api_gateway_authorizers")
func DataSourceAuthorizers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAuthorizersRead,
		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"items": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auth_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authorizer_uri": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"identity_source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authorizer_result_ttl_in_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAuthorizersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	restApiId := d.Get("rest_api_id").(string)

	authorizers, err := conn.GetAuthorizers(&apigateway.GetAuthorizersInput{
		RestApiId: aws.String(restApiId),
	})
	if err != nil {
		return err
	}
	items := make([]map[string]interface{}, 0, len(authorizers.Items))
	for _, item := range authorizers.Items {
		s := make(map[string]interface{})
		s["name"] = item.Name
		s["name"] = item.Name
		s["type"] = item.Type
		s["auth_type"] = item.AuthType
		s["authorizer_uri"] = item.AuthorizerUri
		s["identity_source"] = item.IdentitySource
		s["authorizer_result_ttl_in_seconds"] = item.AuthorizerResultTtlInSeconds
		items = append(items, s)
	}
	if err := d.Set("items", items); err != nil {
		return fmt.Errorf("unable to set authorizer items: %s", err)
	}
	d.SetId(fmt.Sprintf("%s:authorizer", restApiId))
	return nil
}
