package mwaa

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"log"
)

func DataSourceCliToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCliTokenRead,

		Schema: map[string]*schema.Schema{
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"cli_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceCliTokenRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MWAAConn
	environment := d.Get("environment").(string)

	input := &mwaa.CreateCliTokenInput{
		Name: aws.String(environment),
	}

	if v, ok := d.GetOk("environment"); ok {
		input.Name = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Requesting MWAA Cli token: %s", input)
	output, err := conn.CreateCliToken(input)

	if err != nil {
		return fmt.Errorf("requesting MWAA Cli token (%s): %w", environment, err)
	}

	d.SetId(environment)
	d.Set("cli_token", aws.StringValue(output.CliToken))

	return nil

}
