package apigateway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceSdk() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSdkRead,
		Schema: map[string]*schema.Schema{
			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_disposition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sdk_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"java", "javascript", "android", "objectivec", "swift", "ruby"}, false),
			},
			"stage_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceSdkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	restApiId := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	sdkType := d.Get("sdk_type").(string)

	input := &apigateway.GetSdkInput{
		RestApiId: aws.String(restApiId),
		StageName: aws.String(stageName),
		SdkType:   aws.String(sdkType),
	}

	if v, ok := d.GetOk("parameters"); ok && len(v.(map[string]interface{})) > 0 {
		input.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	export, err := conn.GetSdk(input)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", restApiId, stageName, sdkType))
	d.Set("body", string(export.Body))
	d.Set("content_type", export.ContentType)
	d.Set("content_disposition", export.ContentDisposition)

	return nil
}
