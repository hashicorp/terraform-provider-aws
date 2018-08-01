package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/terraform/helper/schema"
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
	log.Print("[DEBUG] Let' go!")
	conn := meta.(*AWSClient).serverlessapplicationrepositoryconn

	applicationID := d.Get("application_id").(string)

	input := &serverlessapplicationrepository.GetApplicationInput{
		ApplicationId: aws.String(applicationID),
	}

	if v, ok := d.GetOk("semantic_version"); ok {
		version := v.(string)
		input.SemanticVersion = aws.String(version)
	}
	log.Printf("[DEBUG] Reading Serverless Repo Application with request: %s", input)

	output, err := conn.GetApplication(input)
	if err != nil {
		return fmt.Errorf("error reading application: %s", err)
	}

	d.SetId(applicationID)
	if err := d.Set("name", output.Name); err != nil {
		return err
	}
	if err := d.Set("semantic_version", output.Version.SemanticVersion); err != nil {
		return err
	}
	if err := d.Set("source_code_url", output.Version.SourceCodeUrl); err != nil {
		return err
	}
	if err := d.Set("template_url", output.Version.TemplateUrl); err != nil {
		return err
	}

	return nil
}
