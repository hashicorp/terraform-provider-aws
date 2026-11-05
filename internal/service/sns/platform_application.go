// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sns

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"maps"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	platformApplicationSchema = map[string]*schema.Schema{
		"apple_platform_bundle_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"apple_platform_team_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		names.AttrARN: {
			Type:     schema.TypeString,
			Computed: true,
		},
		"event_delivery_failure_topic_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"event_endpoint_created_topic_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"event_endpoint_deleted_topic_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"event_endpoint_updated_topic_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"failure_feedback_role_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		names.AttrName: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"platform": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"platform_credential": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			ExactlyOneOf: []string{"platform_credential", "platform_credential_wo"},
		},
		"platform_credential_wo": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			WriteOnly:    true,
			ExactlyOneOf: []string{"platform_credential", "platform_credential_wo"},
			RequiredWith: []string{"platform_credential_wo_version"},
		},
		"platform_credential_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			RequiredWith: []string{"platform_credential_wo"},
		},
		// The principal is optional in the API (only some platforms use it), so the pair uses ConflictsWith rather than ExactlyOneOf.
		"platform_principal": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			ConflictsWith: []string{"platform_principal_wo"},
		},
		"platform_principal_wo": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			WriteOnly:     true,
			ConflictsWith: []string{"platform_principal"},
			RequiredWith:  []string{"platform_principal_wo_version"},
		},
		"platform_principal_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			RequiredWith: []string{"platform_principal_wo"},
		},
		"success_feedback_role_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"success_feedback_sample_rate": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	platformApplicationAttributeMap = attrmap.New(map[string]string{
		"apple_platform_bundle_id":         platformApplicationAttributeNameApplePlatformBundleID,
		"apple_platform_team_id":           platformApplicationAttributeNameApplePlatformTeamID,
		"event_delivery_failure_topic_arn": platformApplicationAttributeNameEventDeliveryFailure,
		"event_endpoint_created_topic_arn": platformApplicationAttributeNameEventEndpointCreated,
		"event_endpoint_deleted_topic_arn": platformApplicationAttributeNameEventEndpointDeleted,
		"event_endpoint_updated_topic_arn": platformApplicationAttributeNameEventEndpointUpdated,
		"failure_feedback_role_arn":        platformApplicationAttributeNameFailureFeedbackRoleARN,
		"platform_credential":              platformApplicationAttributeNamePlatformCredential,
		"platform_principal":               platformApplicationAttributeNamePlatformPrincipal,
		"success_feedback_role_arn":        platformApplicationAttributeNameSuccessFeedbackRoleARN,
		"success_feedback_sample_rate":     platformApplicationAttributeNameSuccessFeedbackSampleRate,
	}, platformApplicationSchema).WithSkipUpdate("apple_platform_bundle_id").WithSkipUpdate("apple_platform_team_id").WithSkipUpdate("platform_credential").WithSkipUpdate("platform_principal")
)

// @SDKResource("aws_sns_platform_application", name="Platform Application")
func resourcePlatformApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlatformApplicationCreate,
		ReadWithoutTimeout:   resourcePlatformApplicationRead,
		UpdateWithoutTimeout: resourcePlatformApplicationUpdate,
		DeleteWithoutTimeout: resourcePlatformApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: maps.Clone(platformApplicationSchema),
	}
}

func resourcePlatformApplicationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	attributes, err := platformApplicationAttributeMap.ResourceDataToAPIAttributesCreate(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	diags = append(diags, expandPlatformApplicationSecretAttributes(d, attributes)...)
	if diags.HasError() {
		return diags
	}

	name := d.Get(names.AttrName).(string)
	input := &sns.CreatePlatformApplicationInput{
		Attributes: attributes,
		Name:       aws.String(name),
		Platform:   aws.String(d.Get("platform").(string)),
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[any, *types.InvalidParameterException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.CreatePlatformApplication(ctx, input)
	}, "is not a valid role to allow SNS to write to Cloudwatch Logs")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS Platform Application (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*sns.CreatePlatformApplicationOutput).PlatformApplicationArn))

	return append(diags, resourcePlatformApplicationRead(ctx, d, meta)...)
}

func resourcePlatformApplicationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	// There is no SNS Describe/GetPlatformApplication to fetch attributes like name and platform
	// We will use the ID, which should be a platform application ARN, to:
	//  * Validate its an appropriate ARN on import
	//  * Parse out the name and platform
	arn, name, platform, err := parsePlatformApplicationResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	attributes, err := findPlatformApplicationAttributesByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SNS Platform Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Platform Application (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, name)
	d.Set("platform", platform)

	err = platformApplicationAttributeMap.APIAttributesToResourceData(attributes, d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourcePlatformApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	attributes, err := platformApplicationAttributeMap.ResourceDataToAPIAttributesUpdate(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("apple_platform_bundle_id", "apple_platform_team_id", "platform_credential", "platform_credential_wo_version", "platform_principal", "platform_principal_wo_version") {
		// If APNS platform was configured with token-based authentication then the only way to update them
		// is to update all 4 attributes as they must be specified together in the request.
		if d.HasChanges("apple_platform_team_id", "apple_platform_bundle_id") {
			attributes[platformApplicationAttributeNameApplePlatformTeamID] = d.Get("apple_platform_team_id").(string)
			attributes[platformApplicationAttributeNameApplePlatformBundleID] = d.Get("apple_platform_bundle_id").(string)
		}

		// Prior to version 3.0.0 of the Terraform AWS Provider, the platform_credential and platform_principal
		// attributes were stored in state as SHA256 hashes. If the changes to these two attributes are the only
		// changes and if both of their changes only match updating the state value, then skip the API call.
		oPCRaw, nPCRaw := d.GetChange("platform_credential")
		oPPRaw, nPPRaw := d.GetChange("platform_principal")

		if len(attributes) == 0 && isChangeSha256Removal(oPCRaw, nPCRaw) && isChangeSha256Removal(oPPRaw, nPPRaw) {
			return diags
		}

		// If the platform requires a principal it must also be specified, even if it didn't change
		// since credential is stored as a hash, the only way to update principal is to update both
		// as they must be specified together in the request.
		diags = append(diags, expandPlatformApplicationSecretAttributes(d, attributes)...)
		if diags.HasError() {
			return diags
		}
	}

	// Make API call to update attributes
	input := &sns.SetPlatformApplicationAttributesInput{
		Attributes:             attributes,
		PlatformApplicationArn: aws.String(d.Id()),
	}

	_, err = tfresource.RetryWhenIsAErrorMessageContains[any, *types.InvalidParameterException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.SetPlatformApplicationAttributes(ctx, input)
	}, "is not a valid role to allow SNS to write to Cloudwatch Logs")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SNS Platform Application (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePlatformApplicationRead(ctx, d, meta)...)
}

func resourcePlatformApplicationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	log.Printf("[DEBUG] Deleting SNS Platform Application: %s", d.Id())
	_, err := conn.DeletePlatformApplication(ctx, &sns.DeletePlatformApplicationInput{
		PlatformApplicationArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SNS Platform Application (%s): %s", d.Id(), err)
	}

	return diags
}

func findPlatformApplicationAttributesByARN(ctx context.Context, conn *sns.Client, arn string) (map[string]string, error) {
	input := &sns.GetPlatformApplicationAttributesInput{
		PlatformApplicationArn: aws.String(arn),
	}

	output, err := conn.GetPlatformApplicationAttributes(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Attributes) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Attributes, nil
}

func parsePlatformApplicationResourceID(input string) (arnS, name, platform string, err error) {
	platformApplicationArn, err := arn.Parse(input)
	if err != nil {
		err = fmt.Errorf(
			"SNS Platform Application ID must be of the form "+
				"arn:PARTITION:sns:REGION:ACCOUNTID:app/PLATFORM/NAME, "+
				"was provided %q and received error: %s", input, err)
		return
	}

	platformApplicationArnResourceParts := strings.Split(platformApplicationArn.Resource, "/")
	if len(platformApplicationArnResourceParts) != 3 || platformApplicationArnResourceParts[0] != "app" {
		err = fmt.Errorf(
			"SNS Platform Application ID must be of the form "+
				"arn:PARTITION:sns:REGION:ACCOUNTID:app/PLATFORM/NAME, "+
				"was provided: %s", input)
		return
	}

	arnS = platformApplicationArn.String()
	name = platformApplicationArnResourceParts[2]
	platform = platformApplicationArnResourceParts[1]
	return
}

// Helper method to add the configured "platform credential" and "principal" secrets to the API attributes map.
// This is split out to encapsulate handling both the regular and write-only variants of these attributes, which
// are mutually exclusive.
func expandPlatformApplicationSecretAttributes(d *schema.ResourceData, attributes map[string]string) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, secret := range []struct {
		apiAttributeName  string
		tfAttributeName   string
		tfAttributeNameWO string
	}{
		{platformApplicationAttributeNamePlatformCredential, "platform_credential", "platform_credential_wo"},
		{platformApplicationAttributeNamePlatformPrincipal, "platform_principal", "platform_principal_wo"},
	} {
		v, di := flex.GetWriteOnlyStringValue(d, cty.GetAttrPath(secret.tfAttributeNameWO))
		diags = append(diags, di...)
		if diags.HasError() {
			return diags
		}

		// If the write-only ("_wo"-suffix) variant isn't set, it's null in the state and so the lookup
		// comes back as an empty string. When this happens, fall back to the "regular" version of the attribute.
		if v == "" {
			v = d.Get(secret.tfAttributeName).(string)
		}
		if v != "" {
			attributes[secret.apiAttributeName] = v
		}
	}

	return diags
}

func isChangeSha256Removal(oldRaw, newRaw any) bool {
	old, ok := oldRaw.(string)
	if !ok {
		return false
	}

	new, ok := newRaw.(string)
	if !ok {
		return false
	}

	hash := sha256.Sum256([]byte(new))
	return hex.EncodeToString(hash[:]) == old
}
