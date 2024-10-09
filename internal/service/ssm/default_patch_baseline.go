// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	patchBaselineIDRegexPattern = `pb-[0-9a-f]{17}`
)

// @SDKResource("aws_ssm_default_patch_baseline", name="Default Patch Baseline")
func resourceDefaultPatchBaseline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultPatchBaselineCreate,
		ReadWithoutTimeout:   resourceDefaultPatchBaselineRead,
		DeleteWithoutTimeout: resourceDefaultPatchBaselineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				id := d.Id()

				if isPatchBaselineID(id) || isPatchBaselineARN(id) {
					conn := meta.(*conns.AWSClient).SSMClient(ctx)

					baseline, err := findPatchBaselineByID(ctx, conn, id)
					if err != nil {
						return nil, fmt.Errorf("reading SSM Patch Baseline (%s): %w", id, err)
					}

					d.SetId(string(baseline.OperatingSystem))
				} else if vals := enum.Values[awstypes.OperatingSystem](); !slices.Contains(vals, id) {
					return nil, fmt.Errorf("ID (%s) must be either a Patch Baseline ID, Patch Baseline ARN, or one of %v", id, vals)
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"baseline_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressPatchBaselineID,
				ValidateFunc: validation.Any(
					validatePatchBaselineID,
					validatePatchBaselineARN,
				),
			},
			"operating_system": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OperatingSystem](),
			},
		},
	}
}

func resourceDefaultPatchBaselineCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	baselineID := d.Get("baseline_id").(string)
	patchBaseline, err := findPatchBaselineByID(ctx, conn, baselineID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Patch Baseline (%s): %s", baselineID, err)
	}

	if pbOS, cOS := patchBaseline.OperatingSystem, awstypes.OperatingSystem(d.Get("operating_system").(string)); pbOS != cOS {
		return sdkdiag.AppendErrorf(diags, "Patch Baseline Operating System (%s) does not match %s", pbOS, cOS)
	}

	input := &ssm.RegisterDefaultPatchBaselineInput{
		BaselineId: aws.String(baselineID),
	}

	_, err = conn.RegisterDefaultPatchBaseline(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering SSM Default Patch Baseline (%s): %s", baselineID, err)
	}

	d.SetId(string(patchBaseline.OperatingSystem))

	return append(diags, resourceDefaultPatchBaselineRead(ctx, d, meta)...)
}

func resourceDefaultPatchBaselineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	output, err := findDefaultPatchBaselineByOperatingSystem(ctx, conn, awstypes.OperatingSystem(d.Id()))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Default Patch Baseline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Default Patch Baseline (%s): %s", d.Id(), err)
	}

	d.Set("baseline_id", output.BaselineId)
	d.Set("operating_system", output.OperatingSystem)

	return diags
}

func resourceDefaultPatchBaselineDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return defaultPatchBaselineRestoreOSDefault(ctx, meta.(*conns.AWSClient).SSMClient(ctx), awstypes.OperatingSystem(d.Id()))
}

func defaultPatchBaselineRestoreOSDefault(ctx context.Context, conn *ssm.Client, os awstypes.OperatingSystem) diag.Diagnostics {
	var diags diag.Diagnostics

	baselineID, err := findDefaultDefaultPatchBaselineIDByOperatingSystem(ctx, conn, os)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AWS-owned SSM Default Patch Baseline for operating system (%s): %s", os, err)
	}

	input := &ssm.RegisterDefaultPatchBaselineInput{
		BaselineId: baselineID,
	}

	_, err = conn.RegisterDefaultPatchBaseline(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "restoring SSM Default Patch Baseline for operating system (%s) to (%s): %s", os, aws.ToString(baselineID), err)
	}

	return diags
}

func findDefaultPatchBaselineByOperatingSystem(ctx context.Context, conn *ssm.Client, os awstypes.OperatingSystem) (*ssm.GetDefaultPatchBaselineOutput, error) {
	input := &ssm.GetDefaultPatchBaselineInput{
		OperatingSystem: os,
	}

	output, err := conn.GetDefaultPatchBaseline(ctx, input)

	if errs.IsA[*awstypes.DoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func patchBaselinesPaginator(conn *ssm.Client, filters ...awstypes.PatchOrchestratorFilter) *ssm.DescribePatchBaselinesPaginator {
	return ssm.NewDescribePatchBaselinesPaginator(conn, &ssm.DescribePatchBaselinesInput{
		Filters: filters,
	})
}

func findDefaultDefaultPatchBaselineIDByOperatingSystem(ctx context.Context, conn *ssm.Client, os awstypes.OperatingSystem) (*string, error) {
	paginator := patchBaselinesPaginator(conn,
		operatingSystemFilter(os),
		ownerIsAWSFilter(),
	)
	re := regexache.MustCompile(`^AWS-[0-9A-Za-z]+PatchBaseline$`)
	var baselineIdentityIDs []string

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("reading SSM Patch Baselines: %s", err)
		}

		for _, v := range page.BaselineIdentities {
			if id := aws.ToString(v.BaselineName); re.MatchString(id) {
				baselineIdentityIDs = append(baselineIdentityIDs, aws.ToString(v.BaselineId))
			}
		}
	}

	return tfresource.AssertSingleValueResult(baselineIdentityIDs)
}

func operatingSystemFilter(os ...awstypes.OperatingSystem) awstypes.PatchOrchestratorFilter {
	return awstypes.PatchOrchestratorFilter{
		Key: aws.String("OPERATING_SYSTEM"),
		Values: tfslices.ApplyToAll(os, func(v awstypes.OperatingSystem) string {
			return string(v)
		}),
	}
}

func ownerIsAWSFilter() awstypes.PatchOrchestratorFilter { // nosemgrep:ci.aws-in-func-name
	return awstypes.PatchOrchestratorFilter{
		Key:    aws.String("OWNER"),
		Values: []string{"AWS"},
	}
}

func ownerIsSelfFilter() awstypes.PatchOrchestratorFilter {
	return awstypes.PatchOrchestratorFilter{
		Key:    aws.String("OWNER"),
		Values: []string{"Self"},
	}
}

func diffSuppressPatchBaselineID(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}

	oldId := oldValue
	if arn.IsARN(oldValue) {
		oldId = patchBaselineIDFromARN(oldValue)
	}

	newId := newValue
	if arn.IsARN(newValue) {
		newId = patchBaselineIDFromARN(newValue)
	}

	if oldId == newId {
		return true
	}

	return false
}

var validatePatchBaselineID = validation.StringMatch(regexache.MustCompile(`^`+patchBaselineIDRegexPattern+`$`), `must match "pb-" followed by 17 hexadecimal characters`)

func validatePatchBaselineARN(v any, k string) (ws []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := arn.Parse(value); err != nil {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid ARN: %s", k, value, err))
		return
	}

	if !isPatchBaselineARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid SSM Patch Baseline ARN", k, value))
		return
	}

	return
}

func isPatchBaselineID(s string) bool {
	re := regexache.MustCompile(`^` + patchBaselineIDRegexPattern + `$`)

	return re.MatchString(s)
}

func isPatchBaselineARN(s string) bool {
	parsedARN, err := arn.Parse(s)
	if err != nil {
		return false
	}

	return patchBaselineIDFromARNResource(parsedARN.Resource) != ""
}

func patchBaselineIDFromARN(s string) string {
	arn, err := arn.Parse(s)
	if err != nil {
		return ""
	}

	return patchBaselineIDFromARNResource(arn.Resource)
}

func patchBaselineIDFromARNResource(s string) string {
	re := regexache.MustCompile(`^patchbaseline/(` + patchBaselineIDRegexPattern + ")$")
	matches := re.FindStringSubmatch(s)
	if matches == nil || len(matches) != 2 {
		return ""
	}

	return matches[1]
}
