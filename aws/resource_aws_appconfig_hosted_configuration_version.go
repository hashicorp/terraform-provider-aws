package aws

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceHostedConfigurationVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostedConfigurationVersionCreate,
		Read:   resourceHostedConfigurationVersionRead,
		Delete: resourceHostedConfigurationVersionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
			},
			"content": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"version_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceHostedConfigurationVersionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	appID := d.Get("application_id").(string)
	profileID := d.Get("configuration_profile_id").(string)

	input := &appconfig.CreateHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(profileID),
		Content:                []byte(d.Get("content").(string)),
		ContentType:            aws.String(d.Get("content_type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateHostedConfigurationVersion(input)

	if err != nil {
		return fmt.Errorf("error creating AppConfig HostedConfigurationVersion for Application (%s): %w", appID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%d", aws.StringValue(output.ApplicationId), aws.StringValue(output.ConfigurationProfileId), aws.Int64Value(output.VersionNumber)))

	return resourceHostedConfigurationVersionRead(d, meta)
}

func resourceHostedConfigurationVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	appID, confProfID, versionNumber, err := resourceAwsAppconfigHostedConfigurationVersionParseID(d.Id())

	if err != nil {
		return err
	}

	input := &appconfig.GetHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
		VersionNumber:          aws.Int64(int64(versionNumber)),
	}

	output, err := conn.GetHostedConfigurationVersion(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appconfig Hosted Configuration Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig Hosted Configuration Version (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig Hosted Configuration Version (%s): empty response", d.Id())
	}

	d.Set("application_id", output.ApplicationId)
	d.Set("configuration_profile_id", output.ConfigurationProfileId)
	d.Set("content", string(output.Content))
	d.Set("content_type", output.ContentType)
	d.Set("description", output.Description)
	d.Set("version_number", output.VersionNumber)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s/hostedconfigurationversion/%d", appID, confProfID, versionNumber),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceHostedConfigurationVersionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	appID, confProfID, versionNumber, err := resourceAwsAppconfigHostedConfigurationVersionParseID(d.Id())

	if err != nil {
		return err
	}

	input := &appconfig.DeleteHostedConfigurationVersionInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
		VersionNumber:          aws.Int64(int64(versionNumber)),
	}

	_, err = conn.DeleteHostedConfigurationVersion(input)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig Hosted Configuration Version (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsAppconfigHostedConfigurationVersionParseID(id string) (string, string, int, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, fmt.Errorf("unexpected format of ID (%q), expected ApplicationID/ConfigurationProfileID/VersionNumber", id)
	}

	version, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", "", 0, fmt.Errorf("error parsing Hosted Configuration Version version_number: %w", err)
	}

	return parts[0], parts[1], version, nil
}
