package configservice

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOrganizationConformancePack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConformancePackCreate,
		ReadWithoutTimeout:   resourceOrganizationConformancePackRead,
		UpdateWithoutTimeout: resourceOrganizationConformancePackUpdate,
		DeleteWithoutTimeout: resourceOrganizationConformancePackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
					ValidateFunc: verify.ValidAccountID,
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
				DiffSuppressFunc: verify.SuppressEquivalentJSONOrYAMLDiffs,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 51200),
					verify.ValidStringIsJSONOrYAML,
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

func resourceOrganizationConformancePackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

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

	_, err := conn.PutOrganizationConformancePackWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Config Organization Conformance Pack (%s): %s", name, err)
	}

	d.SetId(name)

	if err := waitForOrganizationConformancePackStatusCreateSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Config Organization Conformance Pack (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationConformancePackRead(ctx, d, meta)...)
}

func resourceOrganizationConformancePackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	pack, err := DescribeOrganizationConformancePack(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
		log.Printf("[WARN] Config Organization Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Config Organization Conformance Pack (%s): %s", d.Id(), err)
	}

	if pack == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "describing Config Organization Conformance Pack (%s): not found", d.Id())
		}

		log.Printf("[WARN] Config Organization Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", pack.OrganizationConformancePackArn)
	d.Set("name", pack.OrganizationConformancePackName)
	d.Set("delivery_s3_bucket", pack.DeliveryS3Bucket)
	d.Set("delivery_s3_key_prefix", pack.DeliveryS3KeyPrefix)

	if err = d.Set("excluded_accounts", flex.FlattenStringSet(pack.ExcludedAccounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting excluded_accounts: %s", err)
	}

	if err = d.Set("input_parameter", flattenConfigConformancePackInputParameters(pack.ConformancePackInputParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting input_parameter: %s", err)
	}

	return diags
}

func resourceOrganizationConformancePackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

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

	_, err := conn.PutOrganizationConformancePackWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Config Organization Conformance Pack (%s): %s", d.Id(), err)
	}

	if err := waitForOrganizationConformancePackStatusUpdateSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Config Organization Conformance Pack (%s) to be updated: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationConformancePackRead(ctx, d, meta)...)
}

func resourceOrganizationConformancePackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	input := &configservice.DeleteOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationConformancePackWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "erorr deleting Config Organization Conformance Pack (%s): %s", d.Id(), err)
	}

	if err := waitForOrganizationConformancePackStatusDeleteSuccessful(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Config Organization Conformance Pack (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
