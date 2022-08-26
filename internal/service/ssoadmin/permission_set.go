package ssoadmin

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionSetCreate,
		Read:   resourcePermissionSetRead,
		Update: resourcePermissionSetUpdate,
		Delete: resourcePermissionSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 700),
					validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), "must match [\\p{L}\\p{M}\\p{Z}\\p{S}\\p{N}\\p{P}]"),
				),
			},

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				),
			},

			"relay_state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 240),
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9&$@#\\\/%?=~\-_'"|!:,.;*+\[\]\ \(\)\{\}]+`), "must match [a-zA-Z0-9&$@#\\\\\\/%?=~\\-_'\"|!:,.;*+\\[\\]\\(\\)\\{\\}]"),
				),
			},

			"session_duration": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Default:      "PT1H",
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePermissionSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	instanceArn := d.Get("instance_arn").(string)
	name := d.Get("name").(string)

	input := &ssoadmin.CreatePermissionSetInput{
		InstanceArn: aws.String(instanceArn),
		Name:        aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("relay_state"); ok {
		input.RelayState = aws.String(v.(string))
	}

	if v, ok := d.GetOk("session_duration"); ok {
		input.SessionDuration = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreatePermissionSet(input)
	if err != nil {
		return fmt.Errorf("error creating SSO Permission Set (%s): %w", name, err)
	}

	if output == nil || output.PermissionSet == nil {
		return fmt.Errorf("error creating SSO Permission Set (%s): empty output", name)
	}

	d.SetId(fmt.Sprintf("%s,%s", aws.StringValue(output.PermissionSet.PermissionSetArn), instanceArn))

	return resourcePermissionSetRead(d, meta)
}

func resourcePermissionSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn, instanceArn, err := ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Permission Set ID: %w", err)
	}

	output, err := conn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(arn),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] SSO Permission Set (%s) not found, removing from state", arn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SSO Permission Set: %w", err)
	}

	if output == nil || output.PermissionSet == nil {
		return fmt.Errorf("error reading SSO Permission Set (%s): empty output", arn)
	}

	permissionSet := output.PermissionSet

	d.Set("arn", permissionSet.PermissionSetArn)
	d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set("description", permissionSet.Description)
	d.Set("instance_arn", instanceArn)
	d.Set("name", permissionSet.Name)
	d.Set("relay_state", permissionSet.RelayState)
	d.Set("session_duration", permissionSet.SessionDuration)

	tags, err := ListTags(conn, arn, instanceArn)
	if err != nil {
		return fmt.Errorf("error listing tags for SSO Permission Set (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourcePermissionSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	arn, instanceArn, err := ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Permission Set ID: %w", err)
	}

	if d.HasChanges("description", "relay_state", "session_duration") {
		input := &ssoadmin.UpdatePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		}

		// The AWS SSO API requires we send the RelayState value regardless if it's unchanged
		// else the existing Permission Set's RelayState value will be cleared;
		// for consistency, we'll check for the "presence of" instead of "if changed" for all input fields
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17411

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("relay_state"); ok {
			input.RelayState = aws.String(v.(string))
		}

		if v, ok := d.GetOk("session_duration"); ok {
			input.SessionDuration = aws.String(v.(string))
		}

		_, err := conn.UpdatePermissionSet(input)
		if err != nil {
			return fmt.Errorf("error updating SSO Permission Set (%s): %w", arn, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, arn, instanceArn, o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	// Re-provision ALL accounts after making the above changes
	if err := provisionPermissionSet(conn, arn, instanceArn); err != nil {
		return err
	}

	return resourcePermissionSetRead(d, meta)
}

func resourcePermissionSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	arn, instanceArn, err := ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Permission Set ID: %w", err)
	}

	input := &ssoadmin.DeletePermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(arn),
	}

	_, err = conn.DeletePermissionSet(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting SSO Permission Set (%s): %w", arn, err)
	}

	return nil
}

func ParseResourceID(id string) (string, string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%q), expected PERMISSION_SET_ARN,INSTANCE_ARN", id)
	}
	return idParts[0], idParts[1], nil
}

func provisionPermissionSet(conn *ssoadmin.SSOAdmin, arn, instanceArn string) error {
	input := &ssoadmin.ProvisionPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(arn),
		TargetType:       aws.String(ssoadmin.ProvisionTargetTypeAllProvisionedAccounts),
	}

	var output *ssoadmin.ProvisionPermissionSetOutput
	err := resource.Retry(permissionSetProvisionTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.ProvisionPermissionSet(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeConflictException) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeThrottlingException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.ProvisionPermissionSet(input)
	}

	if err != nil {
		return fmt.Errorf("error provisioning SSO Permission Set (%s): %w", arn, err)
	}

	if output == nil || output.PermissionSetProvisioningStatus == nil {
		return fmt.Errorf("error provisioning SSO Permission Set (%s): empty output", arn)
	}

	_, err = waitPermissionSetProvisioned(conn, instanceArn, aws.StringValue(output.PermissionSetProvisioningStatus.RequestId))
	if err != nil {
		return fmt.Errorf("error waiting for SSO Permission Set (%s) to provision: %w", arn, err)
	}

	return nil
}
