package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Read: dataAwsSsmParameterRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"with_decryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"with_default": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"default": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func dataAwsSsmParameterRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	name := d.Get("name").(string)

	paramInput := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(d.Get("with_decryption").(bool)),
	}

	log.Printf("[DEBUG] Reading SSM Parameter: %s", paramInput)
	resp, err := ssmconn.GetParameter(paramInput)

	if err != nil {
		return errwrap.Wrapf("[ERROR] Error describing SSM parameter: {{err}}", err)
	}

	if len(resp.InvalidParameters) > 0 && d.Get("with_default").(bool) == true {
		d.SetId(name)
		d.Set("arn", "")
		d.Set("name", name)
		d.Set("type", "String")
		d.Set("value", d.Get("default").(string))
		return nil
	} else if len(resp.InvalidParameters) > 0 {
		return fmt.Errorf("[ERROR] SSM Parameter %s is invalid", name)
	}

	param := resp.Parameter
	d.SetId(*param.Name)

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "ssm",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("parameter/%s", strings.TrimPrefix(d.Id(), "/")),
	}
	d.Set("arn", arn.String())
	d.Set("name", param.Name)
	d.Set("type", param.Type)
	d.Set("value", param.Value)

	return nil
}
