package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	serverlessrepository "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsServerlessRepositoryApplication() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServerlessRepositoryApplicationRead,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"semantic_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"required_capabilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"source_code_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsServerlessRepositoryApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).serverlessapprepositoryconn

	applicationID := d.Get("application_id").(string)

	input := &serverlessrepository.GetApplicationInput{
		ApplicationId: aws.String(applicationID),
	}

	if v, ok := d.GetOk("semantic_version"); ok {
		version := v.(string)
		input.SemanticVersion = aws.String(version)
	}
	log.Printf("[DEBUG] Reading Serverless Application Repository application with request: %s", input)

	output, err := conn.GetApplication(input)
	if err != nil {
		return fmt.Errorf("error reading Serverless Application Repository application (%s): %w", applicationID, err)
	}

	d.SetId(applicationID)
	d.Set("name", output.Name)
	d.Set("semantic_version", output.Version.SemanticVersion)
	d.Set("source_code_url", output.Version.SourceCodeUrl)
	d.Set("template_url", output.Version.TemplateUrl)
	if err = d.Set("required_capabilities", flattenStringSet(output.Version.RequiredCapabilities)); err != nil {
		return fmt.Errorf("failed to set required_capabilities: %w", err)
	}

	return nil
}
