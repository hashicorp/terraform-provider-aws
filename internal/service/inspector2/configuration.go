// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package inspector2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_inspector2_configuration", name="Configuration")
func resourceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationCreate,
		ReadWithoutTimeout:   resourceConfigurationRead,
		UpdateWithoutTimeout: resourceConfigurationUpdate,
		DeleteWithoutTimeout: resourceConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ec2_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scan_mode": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Ec2ScanMode](),
						},
					},
				},
			},
			"ecr_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rescan_duration": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EcrRescanDuration](),
						},
						"pull_date_rescan_duration": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EcrPullDateRescanDuration](),
						},
						"pull_date_rescan_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EcrPullDateRescanMode](),
						},
					},
				},
			},
		},

		// At least one configuration block must be provided. UpdateConfiguration
		// fails when both ec2_configuration and ecr_configuration are nil.
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
			ec2 := diff.Get("ec2_configuration").([]any)
			ecr := diff.Get("ecr_configuration").([]any)
			if len(ec2) == 0 && len(ecr) == 0 {
				return errConfigurationRequired
			}
			return nil
		},
	}
}

// errConfigurationRequired surfaces the requirement that at least one of
// ec2_configuration or ecr_configuration is set.
var errConfigurationRequired = errors.New("at least one of ec2_configuration or ecr_configuration must be set")

func resourceConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(meta.(*conns.AWSClient).AccountID(ctx))

	return append(diags, resourceConfigurationUpdate(ctx, d, meta)...)
}

func resourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	output, err := findConfiguration(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Inspector2 Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector2 Configuration (%s): %s", d.Id(), err)
	}

	if output.Ec2Configuration != nil && output.Ec2Configuration.ScanModeState != nil {
		if err := d.Set("ec2_configuration", []any{flattenEc2ConfigurationState(output.Ec2Configuration.ScanModeState)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ec2_configuration: %s", err)
		}
	}
	if output.EcrConfiguration != nil && output.EcrConfiguration.RescanDurationState != nil {
		if err := d.Set("ecr_configuration", []any{flattenEcrRescanDurationState(output.EcrConfiguration.RescanDurationState)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ecr_configuration: %s", err)
		}
	}

	return diags
}

func resourceConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	input := &inspector2.UpdateConfigurationInput{}

	if v, ok := d.GetOk("ec2_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Ec2Configuration = expandEc2Configuration(v.([]any)[0].(map[string]any))
	}
	if v, ok := d.GetOk("ecr_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.EcrConfiguration = expandEcrConfiguration(v.([]any)[0].(map[string]any))
	}

	if input.Ec2Configuration == nil && input.EcrConfiguration == nil {
		return sdkdiag.AppendErrorf(diags, "updating Inspector2 Configuration (%s): %s", d.Id(), errConfigurationRequired)
	}

	if _, err := conn.UpdateConfiguration(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Inspector2 Configuration (%s): %s", d.Id(), err)
	}

	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	if _, err := waitConfigurationUpdated(ctx, conn, input, timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Configuration (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceConfigurationRead(ctx, d, meta)...)
}

// resourceConfigurationDelete resets the configuration to AWS defaults.
// The Inspector v2 API has no DeleteConfiguration operation — the configuration
// always exists for an account/region. Defaults align with what AWS provisions
// for newly-onboarded accounts: EC2_HYBRID and ECR LIFETIME with LAST_IN_USE_AT.
func resourceConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	log.Printf("[DEBUG] Resetting Inspector2 Configuration to defaults: %s", d.Id())

	defaults := &inspector2.UpdateConfigurationInput{
		Ec2Configuration: &awstypes.Ec2Configuration{
			ScanMode: awstypes.Ec2ScanModeEc2Hybrid,
		},
		EcrConfiguration: &awstypes.EcrConfiguration{
			RescanDuration:         awstypes.EcrRescanDurationLifetime,
			PullDateRescanDuration: awstypes.EcrPullDateRescanDurationDays90,
			PullDateRescanMode:     awstypes.EcrPullDateRescanModeLastInUseAt,
		},
	}

	if _, err := conn.UpdateConfiguration(ctx, defaults); err != nil {
		return sdkdiag.AppendErrorf(diags, "resetting Inspector2 Configuration (%s): %s", d.Id(), err)
	}

	if _, err := waitConfigurationUpdated(ctx, conn, defaults, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Configuration (%s) reset: %s", d.Id(), err)
	}

	return diags
}

func findConfiguration(ctx context.Context, conn *inspector2.Client) (*inspector2.GetConfigurationOutput, error) {
	input := &inspector2.GetConfigurationInput{}
	output, err := conn.GetConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func waitConfigurationUpdated(ctx context.Context, conn *inspector2.Client, target *inspector2.UpdateConfigurationInput, timeout time.Duration) (*inspector2.GetConfigurationOutput, error) { //nolint:unparam
	var output *inspector2.GetConfigurationOutput

	_, err := tfresource.RetryUntilEqual(ctx, timeout, true, func(ctx context.Context) (bool, error) {
		var err error
		output, err = findConfiguration(ctx, conn)

		if err != nil {
			return false, err
		}

		return configurationMatches(output, target), nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

// configurationMatches reports whether the current state observed via
// GetConfiguration aligns with the requested UpdateConfiguration. Status
// fields are ignored — the API returns PENDING briefly during propagation.
func configurationMatches(state *inspector2.GetConfigurationOutput, target *inspector2.UpdateConfigurationInput) bool {
	if target.Ec2Configuration != nil {
		if state.Ec2Configuration == nil || state.Ec2Configuration.ScanModeState == nil {
			return false
		}
		if state.Ec2Configuration.ScanModeState.ScanMode != target.Ec2Configuration.ScanMode {
			return false
		}
	}
	if target.EcrConfiguration != nil {
		if state.EcrConfiguration == nil || state.EcrConfiguration.RescanDurationState == nil {
			return false
		}
		ecr := state.EcrConfiguration.RescanDurationState
		if ecr.RescanDuration != target.EcrConfiguration.RescanDuration {
			return false
		}
		// Pull-date fields are optional in the input but always returned by the API.
		// Only compare when the user explicitly set them.
		if target.EcrConfiguration.PullDateRescanDuration != "" &&
			ecr.PullDateRescanDuration != target.EcrConfiguration.PullDateRescanDuration {
			return false
		}
		if target.EcrConfiguration.PullDateRescanMode != "" &&
			ecr.PullDateRescanMode != target.EcrConfiguration.PullDateRescanMode {
			return false
		}
	}
	return true
}

func expandEc2Configuration(tfMap map[string]any) *awstypes.Ec2Configuration {
	if tfMap == nil {
		return nil
	}
	out := &awstypes.Ec2Configuration{}
	if v, ok := tfMap["scan_mode"].(string); ok && v != "" {
		out.ScanMode = awstypes.Ec2ScanMode(v)
	}
	return out
}

func expandEcrConfiguration(tfMap map[string]any) *awstypes.EcrConfiguration {
	if tfMap == nil {
		return nil
	}
	out := &awstypes.EcrConfiguration{}
	if v, ok := tfMap["rescan_duration"].(string); ok && v != "" {
		out.RescanDuration = awstypes.EcrRescanDuration(v)
	}
	if v, ok := tfMap["pull_date_rescan_duration"].(string); ok && v != "" {
		out.PullDateRescanDuration = awstypes.EcrPullDateRescanDuration(v)
	}
	if v, ok := tfMap["pull_date_rescan_mode"].(string); ok && v != "" {
		out.PullDateRescanMode = awstypes.EcrPullDateRescanMode(v)
	}
	return out
}

func flattenEc2ConfigurationState(state *awstypes.Ec2ScanModeState) map[string]any {
	if state == nil {
		return nil
	}
	return map[string]any{
		"scan_mode": string(state.ScanMode),
	}
}

func flattenEcrRescanDurationState(state *awstypes.EcrRescanDurationState) map[string]any {
	if state == nil {
		return nil
	}
	return map[string]any{
		"rescan_duration":           string(state.RescanDuration),
		"pull_date_rescan_duration": string(state.PullDateRescanDuration),
		"pull_date_rescan_mode":     string(state.PullDateRescanMode),
	}
}
