// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package accessanalyzer

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
func resourceAnalyzer() *schema.Resource {
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.TypeAccount,
				ValidateDiagFunc: enum.Validate[types.Type](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnalyzerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzerName := d.Get("analyzer_name").(string)
	input := &accessanalyzer.CreateAnalyzerInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(id.UniqueId()),
		Tags:         getTagsIn(ctx),
		Type:         types.Type(d.Get("type").(string)),
	}

	// Handle Organizations eventual consistency.
	_, err := tfresource.RetryWhen(ctx, organizationCreationTimeout,
		func() (interface{}, error) {
			return conn.CreateAnalyzer(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*types.ValidationException](err, "You must create an organization") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Access Analyzer Analyzer (%s): %s", analyzerName, err)
	}

	d.SetId(analyzerName)

	return append(diags, resourceAnalyzerRead(ctx, d, meta)...)
}

func resourceAnalyzerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzer, err := findAnalyzerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Access Analyzer Analyzer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	d.Set("analyzer_name", analyzer.Name)
	d.Set("arn", analyzer.Arn)
	d.Set("type", analyzer.Type)

	setTagsOut(ctx, analyzer.Tags)

	return diags
}

func resourceAnalyzerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAnalyzerRead(ctx, d, meta)...)
}

func resourceAnalyzerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	log.Printf("[DEBUG] Deleting IAM Access Analyzer Analyzer: %s", d.Id())
	_, err := conn.DeleteAnalyzer(ctx, &accessanalyzer.DeleteAnalyzerInput{
		AnalyzerName: aws.String(d.Id()),
		ClientToken:  aws.String(id.UniqueId()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	return diags
}

func findAnalyzerByName(ctx context.Context, conn *accessanalyzer.Client, name string) (*types.AnalyzerSummary, error) {
	input := &accessanalyzer.GetAnalyzerInput{
		AnalyzerName: aws.String(name),
	}

	output, err := conn.GetAnalyzer(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Analyzer == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Analyzer, nil
}
