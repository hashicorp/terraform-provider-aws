package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsSsmParameters() *schema.Resource {
	return &schema.Resource{
		Read: dataAwsSsmParameters,
		Schema: map[string]*schema.Schema{
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"with_decryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"values": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:  true,
				Sensitive: false,
			},
		},
	}
}

func dataAwsSsmParameters(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	path := d.Get("path").(string)
	withDecryption := d.Get("with_decryption").(bool)

	paramInput := &ssm.GetParametersByPathInput{
		Path:           &path,
		WithDecryption: &withDecryption,
	}

	log.Printf("[INFO] Reading SSM Parameters: %s", paramInput)
	resp, err := ssmconn.GetParametersByPath(paramInput)

	if err != nil {
		return fmt.Errorf("Error listing SSM parameters (%s): %w", path, err)
	}

	log.Printf("[INFO] Got %d SSM Parameters", len(resp.Parameters))
	d.SetId(aws.StringValue(paramInput.Path))

	//keys := make([]string, len(resp.Parameters))
	keys := make(map[string]string)
	for _, p := range resp.Parameters {
		log.Printf("[INFO] Got SSM Parameters: %s", *p.Name)
		keys[*p.Name] = *p.Value
	}

	return d.Set("values", keys)
}
