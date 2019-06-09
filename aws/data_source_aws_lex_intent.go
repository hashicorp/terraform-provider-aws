package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceAwsLexIntent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexIntentRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexNameMinLength, lexNameMaxLength),
					validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
				),
			},
			"parent_intent_signature": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "$LATEST",
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
					validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
				),
			},
		},
	}
}

func dataSourceAwsLexIntentRead(d *schema.ResourceData, meta interface{}) error {
	intentName := d.Get("name").(string)
	intentVersion := "$LATEST"
	if v, ok := d.GetOk("version"); ok {
		intentVersion = v.(string)
	}

	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
		Name:    aws.String(intentName),
		Version: aws.String(intentVersion),
	})
	if err != nil {
		return fmt.Errorf("error getting intent %s: %s", intentName, err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.UTC().String())
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.UTC().String())
	d.Set("name", resp.Name)
	d.Set("parent_intent_signature", resp.ParentIntentSignature)
	d.Set("version", resp.Version)

	d.SetId(intentName)

	return nil
}
