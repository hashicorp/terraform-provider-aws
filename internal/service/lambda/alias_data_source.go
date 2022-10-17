package lambda

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliasRead,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"function_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName := d.Get("function_name").(string)
	name := d.Get("name").(string)

	params := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(name),
	}

	aliasConfiguration, err := conn.GetAlias(params)
	if err != nil {
		return fmt.Errorf("Error getting Lambda alias: %w", err)
	}

	d.SetId(aws.StringValue(aliasConfiguration.AliasArn))

	d.Set("arn", aliasConfiguration.AliasArn)
	d.Set("description", aliasConfiguration.Description)
	d.Set("function_version", aliasConfiguration.FunctionVersion)

	invokeArn := functionInvokeARN(*aliasConfiguration.AliasArn, meta)
	d.Set("invoke_arn", invokeArn)

	return nil
}
