package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsAppconfigHostedConfigurationVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppconfigHostedConfigurationVersionCreate,
		Read:   resourceAwsAppconfigHostedConfigurationVersionRead,
		Update: resourceAwsAppconfigHostedConfigurationVersionUpdate,
		Delete: resourceAwsAppconfigHostedConfigurationVersionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppconfigHostedConfigurationVersionImport,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(4, 7),
				),
			},
			"configuration_profile_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(4, 7),
				),
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1024),
				),
			},
			"version_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppconfigHostedConfigurationVersionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	appID := d.Get("application_id").(string)
	profileID := d.Get("configuration_profile_id").(string)

	input := &appconfig.CreateHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(profileID),
		Content:                []byte(d.Get("content").(string)),
		ContentType:            aws.String(d.Get("content_type").(string)),
		Description:            aws.String(d.Get("description").(string)),
	}

	hcv, err := conn.CreateHostedConfigurationVersion(input)
	if err != nil {
		return fmt.Errorf("Error creating AppConfig HostedConfigurationVersion: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%d", appID, profileID, aws.Int64Value(hcv.VersionNumber)))
	d.Set("version_number", hcv.VersionNumber)

	return resourceAwsAppconfigHostedConfigurationVersionRead(d, meta)
}

func resourceAwsAppconfigHostedConfigurationVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.GetHostedConfigurationVersionInput{
		ApplicationId:          aws.String(d.Get("application_id").(string)),
		ConfigurationProfileId: aws.String(d.Get("configuration_profile_id").(string)),
		VersionNumber:          aws.Int64(int64(d.Get("version_number").(int))),
	}

	output, err := conn.GetHostedConfigurationVersion(input)

	if !d.IsNewResource() && isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Appconfig HostedConfigurationVersion (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig HostedConfigurationVersion (%s): %s", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig HostedConfigurationVersion (%s): empty response", d.Id())
	}

	d.Set("description", output.Description)
	d.Set("content", string(output.Content))
	d.Set("content_type", output.ContentType)

	return nil
}

func resourceAwsAppconfigHostedConfigurationVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsAppconfigHostedConfigurationVersionCreate(d, meta)
}

func resourceAwsAppconfigHostedConfigurationVersionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.DeleteHostedConfigurationVersionInput{
		ConfigurationProfileId: aws.String(d.Get("configuration_profile_id").(string)),
		ApplicationId:          aws.String(d.Get("application_id").(string)),
		VersionNumber:          aws.Int64(d.Get("version_number").(int64)),
	}

	_, err := conn.DeleteHostedConfigurationVersion(input)

	if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig HostedConfigurationVersion (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsAppconfigHostedConfigurationVersionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Wrong format of resource: %s. Please follow 'application-id/configurationprofile-id/version-number'", d.Id())
	}

	verString := parts[2]
	verNumber, err := strconv.Atoi(verString)
	if err != nil {
		return nil, fmt.Errorf("version-number must be integer: %s: %w", verString, err)
	}

	d.Set("application_id", parts[0])
	d.Set("configuration_profile_id", parts[1])
	d.Set("version_number", verNumber)

	return []*schema.ResourceData{d}, nil
}
