package ssm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/exp/slices"
)

const (
	patchBaselineIDRegexPattern = `pb-[0-9a-f]{17}`
)

func ResourceDefaultPatchBaseline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultPatchBaselineCreate,
		ReadWithoutTimeout:   resourceDefaultPatchBaselineRead,
		DeleteWithoutTimeout: resourceDefaultPatchBaselineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				id := d.Id()

				if isPatchBaselineID(id) || isPatchBaselineARN(id) {
					conn := meta.(*conns.AWSClient).SSMClient()

					patchbaseline, err := findPatchBaselineByID(ctx, conn, id)
					if err != nil {
						return nil, fmt.Errorf("reading SSM Patch Baseline (%s): %w", id, err)
					}

					d.SetId(string(patchbaseline.OperatingSystem))
				} else if vals := enum.Values[types.OperatingSystem](); !slices.Contains(vals, id) {
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
				ValidateDiagFunc: enum.Validate[types.OperatingSystem](),
			},
		},
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

var validatePatchBaselineID = validation.StringMatch(regexp.MustCompile(`^`+patchBaselineIDRegexPattern+`$`), `must match "pb-" followed by 17 hexadecimal characters`)

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
	re := regexp.MustCompile(`^` + patchBaselineIDRegexPattern + `$`)

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
	re := regexp.MustCompile(`^patchbaseline/(` + patchBaselineIDRegexPattern + ")$")
	matches := re.FindStringSubmatch(s)
	if matches == nil || len(matches) != 2 {
		return ""
	}

	return matches[1]
}

const (
	ResNameDefaultPatchBaseline = "Default Patch Baseline"
)

func resourceDefaultPatchBaselineCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMClient()

	baselineID := d.Get("baseline_id").(string)

	patchBaseline, err := findPatchBaselineByID(ctx, conn, baselineID)
	if err != nil {
		return create.DiagErrorMessage(names.SSM, "registering", ResNameDefaultPatchBaseline, baselineID,
			create.ProblemStandardMessage(names.SSM, create.ErrActionReading, resNamePatchBaseline, baselineID, err),
		)
	}
	if pbOS, cOS := string(patchBaseline.OperatingSystem), d.Get("operating_system"); pbOS != cOS {
		return create.DiagErrorMessage(names.SSM, "registering", ResNameDefaultPatchBaseline, baselineID,
			fmt.Sprintf("Patch Baseline Operating System (%s) does not match %s", pbOS, cOS),
		)
	}

	in := &ssm.RegisterDefaultPatchBaselineInput{
		BaselineId: aws.String(baselineID),
	}
	_, err = conn.RegisterDefaultPatchBaseline(ctx, in)
	if err != nil {
		return create.DiagError(names.SSM, "registering", ResNameDefaultPatchBaseline, baselineID, err)
	}

	// We need to retrieve the Operating System from the Patch Baseline to store for the ID

	d.SetId(string(patchBaseline.OperatingSystem))

	return resourceDefaultPatchBaselineRead(ctx, d, meta)
}

func resourceDefaultPatchBaselineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMClient()

	out, err := FindDefaultPatchBaseline(ctx, conn, types.OperatingSystem(d.Id()))
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Default Patch Baseline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return create.DiagError(names.SSM, create.ErrActionReading, ResNameDefaultPatchBaseline, d.Id(), err)
	}

	d.Set("baseline_id", out.BaselineId)
	d.Set("operating_system", out.OperatingSystem)

	return nil
}

func operatingSystemFilter(os ...string) types.PatchOrchestratorFilter {
	return types.PatchOrchestratorFilter{
		Key:    aws.String("OPERATING_SYSTEM"),
		Values: os,
	}
}

func ownerIsAWSFilter() types.PatchOrchestratorFilter { // nosemgrep:ci.aws-in-func-name
	return types.PatchOrchestratorFilter{
		Key:    aws.String("OWNER"),
		Values: []string{"AWS"},
	}
}

func resourceDefaultPatchBaselineDelete(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	return defaultPatchBaselineRestoreOSDefault(ctx, meta.(*conns.AWSClient), d.Id())
}

func defaultPatchBaselineRestoreOSDefault(ctx context.Context, meta *conns.AWSClient, os string) (diags diag.Diagnostics) {
	conn := meta.SSMClient()

	baselineID, err := FindDefaultDefaultPatchBaselineIDForOS(ctx, conn, types.OperatingSystem(os))
	if errors.Is(err, tfresource.ErrEmptyResult) {
		diags = errs.AppendWarningf(diags, "no AWS-owned default Patch Baseline found for operating system %q", os)
		return
	}
	var tmr *tfresource.TooManyResultsError
	if errors.As(err, &tmr) {
		diags = errs.AppendWarningf(diags, "found %d AWS-owned default Patch Baselines found for operating system %q", tmr.Count, os)
	}

	in := &ssm.RegisterDefaultPatchBaselineInput{
		BaselineId: aws.String(baselineID),
	}
	_, err = conn.RegisterDefaultPatchBaseline(ctx, in)
	if err != nil {
		diags = errs.AppendErrorf(diags, "restoring SSM Default Patch Baseline for operating system %q to %q: %s", os, baselineID, err)
	}

	return
}

func FindDefaultPatchBaseline(ctx context.Context, conn *ssm.Client, os types.OperatingSystem) (*ssm.GetDefaultPatchBaselineOutput, error) {
	in := &ssm.GetDefaultPatchBaselineInput{
		OperatingSystem: os,
	}
	out, err := conn.GetDefaultPatchBaseline(ctx, in)
	if err != nil {
		var nfe *types.DoesNotExistException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func findPatchBaselineByID(ctx context.Context, conn *ssm.Client, id string) (*ssm.GetPatchBaselineOutput, error) {
	in := &ssm.GetPatchBaselineInput{
		BaselineId: aws.String(id),
	}
	out, err := conn.GetPatchBaseline(ctx, in)
	if err != nil {
		var nfe *types.DoesNotExistException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func patchBaselinesPaginator(conn *ssm.Client, filters ...types.PatchOrchestratorFilter) *ssm.DescribePatchBaselinesPaginator {
	return ssm.NewDescribePatchBaselinesPaginator(conn, &ssm.DescribePatchBaselinesInput{
		Filters: filters,
	})
}

func FindDefaultDefaultPatchBaselineIDForOS(ctx context.Context, conn *ssm.Client, os types.OperatingSystem) (string, error) {
	paginator := patchBaselinesPaginator(conn,
		operatingSystemFilter(string(os)),
		ownerIsAWSFilter(),
	)
	re := regexp.MustCompile(`^AWS-[A-Za-z0-9]+PatchBaseline$`)
	var baselineIdentityIDs []string
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("listing Patch Baselines for operating system %q: %s", os, err)
		}

		for _, identity := range out.BaselineIdentities {
			if id := aws.ToString(identity.BaselineName); re.MatchString(id) {
				baselineIdentityIDs = append(baselineIdentityIDs, aws.ToString(identity.BaselineId))
			}
		}
	}

	if l := len(baselineIdentityIDs); l == 0 {
		return "", tfresource.NewEmptyResultError(nil)
	} else if l > 1 {
		return "", tfresource.NewTooManyResultsError(l, nil)
	}

	return baselineIdentityIDs[0], nil
}
