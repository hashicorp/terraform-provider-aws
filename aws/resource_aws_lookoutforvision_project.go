package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lookoutforvision"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lookoutforvision/finder"
)

func resourceAwsLookoutForVisionProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLookoutForVisionProjectCreate,
		Read:   resourceAwsLookoutForVisionProjectRead,
		Delete: resourceAwsLookoutForVisionProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](_*-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, - (hyphen) and _ (underscore). Name must begin with an alphanumeric character."),
				),
			},
		},
	}
}

func resourceAwsLookoutForVisionProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lookoutforvisionconn

	name := d.Get("name").(string)

	input := &lookoutforvision.CreateProjectInput{
		ProjectName: aws.String(name),
		ClientToken: aws.String(resource.UniqueId()),
	}

	log.Printf("[DEBUG] Amazon Lookout for Vision project create config: %#v", *input)
	_, err := conn.CreateProject(input)
	if err != nil {
		return fmt.Errorf("Error creating Amazon Lookout for Vision project: %w", err)
	}

	d.SetId(name)

	return resourceAwsLookoutForVisionProjectRead(d, meta)
}

func resourceAwsLookoutForVisionProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lookoutforvisionconn

	project, err := finder.ProjectByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ValidationException", "Cannot find project") {
			d.SetId("")
			log.Printf("[WARN] Unable to find Amazon Lookout for Vision project (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading Amazon Lookout for Vision project (%s): %w", d.Id(), err)

	}

	d.Set("name", project.ProjectDescription.ProjectName)
	d.Set("arn", project.ProjectDescription.ProjectArn)

	return nil
}

func resourceAwsLookoutForVisionProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lookoutforvisionconn

	input := &lookoutforvision.DeleteProjectInput{
		ProjectName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteProject(input); err != nil {
		if isAWSErr(err, "ValidationException", "Cannot find project") {
			return nil
		}
		return fmt.Errorf("error deleting Lookout for Vision project (%s): %w", d.Id(), err)
	}

	return nil
}
