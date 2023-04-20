package accessanalyzer

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for Organizations eventual consistency on creation
	// This timeout value is much higher than usual since the cross-service validation
	// appears to be consistently caching for 5 minutes:
	// --- PASS: TestAccAccessAnalyzer_serial/Analyzer/Type_Organization (315.86s)
	organizationCreationTimeout = 10 * time.Minute
)

// @SDKResource("aws_accessanalyzer_analyzer", name="Analyzer")
// @Tags(identifierAttribute="arn")
func ResourceAnalyzer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnalyzerCreate,
		ReadWithoutTimeout:   resourceAnalyzerRead,
		UpdateWithoutTimeout: resourceAnalyzerUpdate,
		DeleteWithoutTimeout: resourceAnalyzerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"analyzer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]*$`), "must begin with a letter and contain only alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  accessanalyzer.TypeAccount,
				ValidateFunc: validation.StringInSlice([]string{
					accessanalyzer.TypeAccount,
					accessanalyzer.TypeOrganization,
				}, false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnalyzerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn()

	analyzerName := d.Get("analyzer_name").(string)
	input := &accessanalyzer.CreateAnalyzerInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(id.UniqueId()),
		Tags:         GetTagsIn(ctx),
		Type:         aws.String(d.Get("type").(string)),
	}

	// Handle Organizations eventual consistency
	err := retry.RetryContext(ctx, organizationCreationTimeout, func() *retry.RetryError {
		_, err := conn.CreateAnalyzerWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, accessanalyzer.ErrCodeValidationException, "You must create an organization") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateAnalyzerWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Access Analyzer Analyzer (%s): %s", analyzerName, err)
	}

	d.SetId(analyzerName)

	return append(diags, resourceAnalyzerRead(ctx, d, meta)...)
}

func resourceAnalyzerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn()

	input := &accessanalyzer.GetAnalyzerInput{
		AnalyzerName: aws.String(d.Id()),
	}

	output, err := conn.GetAnalyzerWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, accessanalyzer.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Access Analyzer Analyzer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	if output == nil || output.Analyzer == nil {
		return sdkdiag.AppendErrorf(diags, "getting Access Analyzer Analyzer (%s): empty response", d.Id())
	}

	d.Set("analyzer_name", output.Analyzer.Name)
	d.Set("arn", output.Analyzer.Arn)

	SetTagsOut(ctx, output.Analyzer.Tags)

	d.Set("type", output.Analyzer.Type)

	return diags
}

func resourceAnalyzerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAnalyzerRead(ctx, d, meta)...)
}

func resourceAnalyzerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn()

	log.Printf("[DEBUG] Deleting Access Analyzer Analyzer: (%s)", d.Id())
	_, err := conn.DeleteAnalyzerWithContext(ctx, &accessanalyzer.DeleteAnalyzerInput{
		AnalyzerName: aws.String(d.Id()),
		ClientToken:  aws.String(id.UniqueId()),
	})

	if tfawserr.ErrCodeEquals(err, accessanalyzer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	return diags
}
