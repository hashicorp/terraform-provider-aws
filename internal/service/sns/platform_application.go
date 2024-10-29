// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
		"platform_principal": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
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

// @SDKResource("aws_sns_platform_application")
func resourcePlatformApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlatformApplicationCreate,
		ReadWithoutTimeout:   resourcePlatformApplicationRead,
		UpdateWithoutTimeout: resourcePlatformApplicationUpdate,
		DeleteWithoutTimeout: resourcePlatformApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: platformApplicationSchema,
	}
}

func resourcePlatformApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	attributes, err := platformApplicationAttributeMap.ResourceDataToAPIAttributesCreate(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get(names.AttrName).(string)
	input := &sns.CreatePlatformApplicationInput{
		Attributes: attributes,
		Name:       aws.String(name),
		Platform:   aws.String(d.Get("platform").(string)),
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreatePlatformApplication(ctx, input)
	}, "is not a valid role to allow SNS to write to Cloudwatch Logs")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS Platform Application (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*sns.CreatePlatformApplicationOutput).PlatformApplicationArn))

	return append(diags, resourcePlatformApplicationRead(ctx, d, meta)...)
}

func resourcePlatformApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

func resourcePlatformApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	attributes, err := platformApplicationAttributeMap.ResourceDataToAPIAttributesUpdate(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("apple_platform_bundle_id", "apple_platform_team_id", "platform_credential", "platform_principal") {
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

		attributes[platformApplicationAttributeNamePlatformCredential] = d.Get("platform_credential").(string)
		// If the platform requires a principal it must also be specified, even if it didn't change
		// since credential is stored as a hash, the only way to update principal is to update both
		// as they must be specified together in the request.
		if v, ok := d.GetOk("platform_principal"); ok {
			attributes[platformApplicationAttributeNamePlatformPrincipal] = v.(string)
		}
	}

	// Make API call to update attributes
	input := &sns.SetPlatformApplicationAttributesInput{
		Attributes:             attributes,
		PlatformApplicationArn: aws.String(d.Id()),
	}

	_, err = tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.SetPlatformApplicationAttributes(ctx, input)
	}, "is not a valid role to allow SNS to write to Cloudwatch Logs")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SNS Platform Application (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePlatformApplicationRead(ctx, d, meta)...)
}

func resourcePlatformApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Attributes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
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

func isChangeSha256Removal(oldRaw, newRaw interface{}) bool {
	old, ok := oldRaw.(string)
	if !ok {
		return false
	}

	new, ok := newRaw.(string)
	if !ok {
		return false
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(new))) == old
}
