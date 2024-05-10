// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"errors"
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

// @SDKResource("aws_securityhub_configuration_policy_association", name="Configuration Policy Association")
func resourceConfigurationPolicyAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationPolicyAssociationCreateOrUpdate,
		ReadWithoutTimeout:   resourceConfigurationPolicyAssociationRead,
		UpdateWithoutTimeout: resourceConfigurationPolicyAssociationCreateOrUpdate,
		DeleteWithoutTimeout: resourceConfigurationPolicyAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Second),
			Update: schema.DefaultTimeout(90 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The universally unique identifier (UUID) of the configuration policy.",
				ValidateFunc: validation.IsUUID,
			},
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
		},
	}
}

func resourceConfigurationPolicyAssociationCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	targetID := d.Get("target_id").(string)
	input := &securityhub.StartConfigurationPolicyAssociationInput{
		ConfigurationPolicyIdentifier: aws.String(d.Get("policy_id").(string)),
		Target:                        expandTarget(targetID),
	}

	_, err := conn.StartConfigurationPolicyAssociation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting Security Hub Configuration Policy Association (%s): %s", targetID, err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	if d.IsNewResource() {
		d.SetId(targetID)
	} else {
		timeout = d.Timeout(schema.TimeoutUpdate)
	}

	if _, err := waitConfigurationPolicyAssociationSucceeded(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Configuration Policy Association (%s) success: %s", targetID, err)
	}

	return append(diags, resourceConfigurationPolicyAssociationRead(ctx, d, meta)...)
}

func resourceConfigurationPolicyAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := findConfigurationPolicyAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Configuration Policy Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Configuration Policy Association (%s): %s", d.Id(), err)
	}

	d.Set("policy_id", output.ConfigurationPolicyId)
	d.Set("target_id", output.TargetId)

	return diags
}

func resourceConfigurationPolicyAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Configuration Policy Association: %s", d.Id())
	_, err := conn.StartConfigurationPolicyDisassociation(ctx, &securityhub.StartConfigurationPolicyDisassociationInput{
		ConfigurationPolicyIdentifier: aws.String(d.Get("policy_id").(string)),
		Target:                        expandTarget(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting Security Hub Configuration Policy Disassociation (%s): %s", d.Id(), err)
	}

	return diags
}

func findConfigurationPolicyAssociationByID(ctx context.Context, conn *securityhub.Client, id string) (*securityhub.GetConfigurationPolicyAssociationOutput, error) {
	input := &securityhub.GetConfigurationPolicyAssociationInput{
		Target: expandTarget(id),
	}

	return findConfigurationPolicyAssociation(ctx, conn, input)
}

func findConfigurationPolicyAssociation(ctx context.Context, conn *securityhub.Client, input *securityhub.GetConfigurationPolicyAssociationInput) (*securityhub.GetConfigurationPolicyAssociationOutput, error) {
	output, err := conn.GetConfigurationPolicyAssociation(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "Must be a Security Hub delegated administrator with Central Configuration enabled") || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
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

func statusConfigurationPolicyAssociation(ctx context.Context, conn *securityhub.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConfigurationPolicyAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AssociationStatus), nil
	}
}

func waitConfigurationPolicyAssociationSucceeded(ctx context.Context, conn *securityhub.Client, id string, timeout time.Duration) (*securityhub.GetConfigurationPolicyAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConfigurationPolicyAssociationStatusPending),
		Target:  enum.Slice(types.ConfigurationPolicyAssociationStatusSuccess),
		Refresh: statusConfigurationPolicyAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if tfresource.TimedOut(err) {
		log.Printf("[WARN] Security Hub Configuration Policy Association (%s) still in PENDING state. It can take up to 24 hours for the status to change from PENDING to SUCCESS or FAILURE", id)
		// We try to wait until SUCCESS state is reached but don't error if still in PENDING state.
		// We must attempt to wait/retry in order for Policy Disassociations to take effect
		return findConfigurationPolicyAssociationByID(ctx, conn, id)
	}

	if output, ok := outputRaw.(*securityhub.GetConfigurationPolicyAssociationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.AssociationStatusMessage)))

		return output, err
	}

	return nil, err
}

func expandTarget(targetID string) types.Target {
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
