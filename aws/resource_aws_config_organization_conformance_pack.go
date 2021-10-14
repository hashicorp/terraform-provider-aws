package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func resourceAwsConfigOrganizationConformancePack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigOrganizationConformancePackCreate,
		Read:   resourceAwsConfigOrganizationConformancePackRead,
		Update: resourceAwsConfigOrganizationConformancePackUpdate,
		Delete: resourceAwsConfigOrganizationConformancePackDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_s3_bucket": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^awsconfigconforms`), `must begin with "awsconfigconforms"`),
				),
			},
			"delivery_s3_key_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"excluded_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1000,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateAwsAccountId,
				},
			},
			"input_parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 60,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameter_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"parameter_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"template_body": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJsonOrYamlDiffs,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 51200),
					validateStringIsJsonOrYaml,
				),
				ConflictsWith: []string{"template_s3_uri"},
			},
			"template_s3_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`^s3://`), "must begin with s3://"),
				),
				ConflictsWith: []string{"template_body"},
			},
		},
	}
}

func resourceAwsConfigOrganizationConformancePackCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigConn

	name := d.Get("name").(string)

	input := &configservice.PutOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(name),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}

	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok {
		input.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameter"); ok {
		input.ConformancePackInputParameters = expandConfigConformancePackInputParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_s3_uri"); ok {
		input.TemplateS3Uri = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConformancePack(input)

	if err != nil {
		return fmt.Errorf("error creating Config Organization Conformance Pack (%s): %w", name, err)
	}

	d.SetId(name)

	if err := configWaitForOrganizationConformancePackStatusCreateSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Conformance Pack (%s) to be created: %w", d.Id(), err)
	}

	return resourceAwsConfigOrganizationConformancePackRead(d, meta)
}

func resourceAwsConfigOrganizationConformancePackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigConn

	pack, err := configDescribeOrganizationConformancePack(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
		log.Printf("[WARN] Config Organization Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Config Organization Conformance Pack (%s): %w", d.Id(), err)
	}

	if pack == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error describing Config Organization Conformance Pack (%s): not found", d.Id())
		}

		log.Printf("[WARN] Config Organization Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", pack.OrganizationConformancePackArn)
	d.Set("name", pack.OrganizationConformancePackName)
	d.Set("delivery_s3_bucket", pack.DeliveryS3Bucket)
	d.Set("delivery_s3_key_prefix", pack.DeliveryS3KeyPrefix)

	if err = d.Set("excluded_accounts", flex.FlattenStringSet(pack.ExcludedAccounts)); err != nil {
		return fmt.Errorf("error setting excluded_accounts: %w", err)
	}

	if err = d.Set("input_parameter", flattenConfigConformancePackInputParameters(pack.ConformancePackInputParameters)); err != nil {
		return fmt.Errorf("error setting input_parameter: %w", err)
	}

	return nil
}

func resourceAwsConfigOrganizationConformancePackUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigConn

	input := &configservice.PutOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}

	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok {
		input.ExcludedAccounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameter"); ok {
		input.ConformancePackInputParameters = expandConfigConformancePackInputParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_s3_uri"); ok {
		input.TemplateS3Uri = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConformancePack(input)

	if err != nil {
		return fmt.Errorf("error updating Config Organization Conformance Pack (%s): %w", d.Id(), err)
	}

	if err := configWaitForOrganizationConformancePackStatusUpdateSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Conformance Pack (%s) to be updated: %w", d.Id(), err)
	}

	return resourceAwsConfigOrganizationConformancePackRead(d, meta)
}

func resourceAwsConfigOrganizationConformancePackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigConn

	input := &configservice.DeleteOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationConformancePack(input)

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("erorr deleting Config Organization Conformance Pack (%s): %w", d.Id(), err)
	}

	if err := configWaitForOrganizationConformancePackStatusDeleteSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
			return nil
		}
		return fmt.Errorf("error waiting for Config Organization Conformance Pack (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
