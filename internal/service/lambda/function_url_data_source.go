package lambda

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"log"
)

func DataSourceFunctionUrl() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFunctionUrlRead,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authorization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_credentials": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"allow_headers": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_methods": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_origins": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFunctionUrlRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	input := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Getting Lambda Function Url Config: %s", input)
	output, err := conn.GetFunctionUrlConfig(input)

	if err != nil {
		return fmt.Errorf("error getting Lambda Function Url Config: %w", err)
	}

	d.SetId(aws.StringValue(output.FunctionArn))
	if err = d.Set("authorization_type", output.AuthorizationType); err != nil {
		return err
	}
	if err = d.Set("cors", flattenFunctionUrlCorsConfigs(output.Cors)); err != nil {
		return err
	}
	if err = d.Set("creation_time", output.CreationTime); err != nil {
		return err
	}
	if err = d.Set("function_arn", output.FunctionArn); err != nil {
		return err
	}
	if err = d.Set("function_url", output.FunctionUrl); err != nil {
		return err
	}
	if err = d.Set("last_modified_time", output.LastModifiedTime); err != nil {
		return err
	}

	return nil
}
