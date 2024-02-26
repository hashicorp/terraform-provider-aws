// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_securityhub_configuration_policy_association")
func ResourceConfigurationPolicyAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationPolicyAssociationCreateOrUpdate,
		ReadWithoutTimeout:   resourceConfigurationPolicyAssociationRead,
		UpdateWithoutTimeout: resourceConfigurationPolicyAssociationCreateOrUpdate,
		DeleteWithoutTimeout: resourceConfigurationPolicyAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(90 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier of the target account, organizational unit, or the root to associate with the specified configuration.",
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`^(r-[a-z0-9]{4,32})$|^(ou-[a-z0-9]{4,32}-[a-z0-9]{8,32})$|^([0-9]{12})$`),
					"Target ID must be a valid root, organizational unit or account id.",
				),
			},
			"policy_id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The universally unique identifier (UUID) of the configuration policy.",
				ValidateFunc: validation.IsUUID,
			},
		},
	}
}

// GetTarget converts a target id string into proper types.Target struct.
func GetTarget(targetID string) types.Target {
	if strings.HasPrefix(targetID, "r-") {
		return &types.TargetMemberRootId{
			Value: targetID,
		}
	}

	if strings.HasPrefix(targetID, "ou-") {
		return &types.TargetMemberOrganizationalUnitId{
			Value: targetID,
		}
	}

	return &types.TargetMemberAccountId{
		Value: targetID,
	}
}

func resourceConfigurationPolicyAssociationCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	targetID := d.Get("target_id").(string)
	input := &securityhub.StartConfigurationPolicyAssociationInput{
		ConfigurationPolicyIdentifier: aws.String(d.Get("policy_id").(string)),
		Target:                        GetTarget(targetID),
	}

	_, err := conn.StartConfigurationPolicyAssociation(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting Security Hub Configuration Policy Association (%s): %s", targetID, err)
	}

	if d.IsNewResource() {
		d.SetId(targetID)
	}

	return append(diags, resourceConfigurationPolicyAssociationRead(ctx, d, meta)...)
}

func resourceConfigurationPolicyAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.StartConfigurationPolicyDisassociationInput{
		ConfigurationPolicyIdentifier: aws.String(d.Get("policy_id").(string)),
		Target:                        GetTarget(d.Get("target_id").(string)),
	}

	_, err := conn.StartConfigurationPolicyDisassociation(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting Security Hub Configuration Policy Disassociation (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceConfigurationPolicyAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	out, err := waitConfigurationPolicyAssociationSuccess(ctx, conn, GetTarget(d.Id()), d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Configuration Policy Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Configuration Policy Association (%s): %s", d.Id(), err)
	}

	d.Set("policy_id", out.ConfigurationPolicyId)
	d.Set("target_id", out.TargetId)
	return diags
}

func waitConfigurationPolicyAssociationSuccess(ctx context.Context, conn *securityhub.Client, target types.Target, timeout time.Duration) (*securityhub.GetConfigurationPolicyAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ConfigurationPolicyAssociationStatusPending),
		Target:                    enum.Slice(types.ConfigurationPolicyAssociationStatusSuccess),
		Refresh:                   getConfigurationPolicyAssociation(ctx, conn, target),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	var timeoutErr *retry.TimeoutError
	if tfresource.TimedOut(err) && errors.As(err, &timeoutErr) {
		log.Printf("[WARN] Security Hub Configuration Policy Association still in state: %s. It can take up to 24 hours for the status to change from PENDING to SUCCESS or FAILURE", timeoutErr.LastState)
		// We try to wait until SUCCESS state is reached but don't error if still in PENDING state.
		// We must attempt to wait/retry in order for Policy Disassociations to take effect
		return conn.GetConfigurationPolicyAssociation(ctx, &securityhub.GetConfigurationPolicyAssociationInput{
			Target: target,
		})
	}

	return outputRaw.(*securityhub.GetConfigurationPolicyAssociationOutput), err
}

func getConfigurationPolicyAssociation(ctx context.Context, conn *securityhub.Client, target types.Target) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &securityhub.GetConfigurationPolicyAssociationInput{
			Target: target,
		}
		output, err := conn.GetConfigurationPolicyAssociation(ctx, input)
		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, "", &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.TargetId == nil {
			return nil, "", tfresource.NewEmptyResultError(input)
		}

		switch output.AssociationStatus {
		case types.ConfigurationPolicyAssociationStatusPending:
			return output, string(output.AssociationStatus), nil
		case types.ConfigurationPolicyAssociationStatusSuccess, "":
			return output, string(output.AssociationStatus), nil
		default:
			var statusErr error
			if msg := output.AssociationStatusMessage; msg != nil && len(*msg) > 0 {
				statusErr = fmt.Errorf("StatusMessage: %s", *msg)
			}
			return nil, "", &retry.UnexpectedStateError{
				LastError: statusErr,
				State:     string(output.AssociationStatus),
				ExpectedState: []string{
					string(types.ConfigurationPolicyAssociationStatusPending),
					string(types.ConfigurationPolicyAssociationStatusSuccess),
				},
			}
		}
	}
}
