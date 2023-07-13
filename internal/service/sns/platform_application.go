// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
		"arn": {
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
		"name": {
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
		"apple_platform_bundle_id":         PlatformApplicationAttributeNameApplePlatformBundleID,
		"apple_platform_team_id":           PlatformApplicationAttributeNameApplePlatformTeamID,
		"event_delivery_failure_topic_arn": PlatformApplicationAttributeNameEventDeliveryFailure,
		"event_endpoint_created_topic_arn": PlatformApplicationAttributeNameEventEndpointCreated,
		"event_endpoint_deleted_topic_arn": PlatformApplicationAttributeNameEventEndpointDeleted,
		"event_endpoint_updated_topic_arn": PlatformApplicationAttributeNameEventEndpointUpdated,
		"failure_feedback_role_arn":        PlatformApplicationAttributeNameFailureFeedbackRoleARN,
		"platform_credential":              PlatformApplicationAttributeNamePlatformCredential,
		"platform_principal":               PlatformApplicationAttributeNamePlatformPrincipal,
		"success_feedback_role_arn":        PlatformApplicationAttributeNameSuccessFeedbackRoleARN,
		"success_feedback_sample_rate":     PlatformApplicationAttributeNameSuccessFeedbackSampleRate,
	}, platformApplicationSchema).WithSkipUpdate("apple_platform_bundle_id").WithSkipUpdate("apple_platform_team_id").WithSkipUpdate("platform_credential").WithSkipUpdate("platform_principal")
)

// @SDKResource("aws_sns_platform_application")
func ResourcePlatformApplication() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	attributes, err := platformApplicationAttributeMap.ResourceDataToAPIAttributesCreate(d)

	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	input := &sns.CreatePlatformApplicationInput{
		Attributes: aws.StringMap(attributes),
		Name:       aws.String(name),
		Platform:   aws.String(d.Get("platform").(string)),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreatePlatformApplicationWithContext(ctx, input)
	}, sns.ErrCodeInvalidParameterException, "is not a valid role to allow SNS to write to Cloudwatch Logs")

	if err != nil {
		return diag.Errorf("creating SNS Platform Application (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*sns.CreatePlatformApplicationOutput).PlatformApplicationArn))

	return resourcePlatformApplicationRead(ctx, d, meta)
}

func resourcePlatformApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	// There is no SNS Describe/GetPlatformApplication to fetch attributes like name and platform
	// We will use the ID, which should be a platform application ARN, to:
	//  * Validate its an appropriate ARN on import
	//  * Parse out the name and platform
	arn, name, platform, err := DecodePlatformApplicationID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	attributes, err := FindPlatformApplicationAttributesByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Platform Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading SNS Platform Application (%s): %s", d.Id(), err)
	}

	d.Set("arn", arn)
	d.Set("name", name)
	d.Set("platform", platform)

	err = platformApplicationAttributeMap.APIAttributesToResourceData(attributes, d)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePlatformApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	attributes, err := platformApplicationAttributeMap.ResourceDataToAPIAttributesUpdate(d)

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("apple_platform_bundle_id", "apple_platform_team_id", "platform_credential", "platform_principal") {
		// If APNS platform was configured with token-based authentication then the only way to update them
		// is to update all 4 attributes as they must be specified together in the request.
		if d.HasChanges("apple_platform_team_id", "apple_platform_bundle_id") {
			attributes[PlatformApplicationAttributeNameApplePlatformTeamID] = d.Get("apple_platform_team_id").(string)
			attributes[PlatformApplicationAttributeNameApplePlatformBundleID] = d.Get("apple_platform_bundle_id").(string)
		}

		// Prior to version 3.0.0 of the Terraform AWS Provider, the platform_credential and platform_principal
		// attributes were stored in state as SHA256 hashes. If the changes to these two attributes are the only
		// changes and if both of their changes only match updating the state value, then skip the API call.
		oPCRaw, nPCRaw := d.GetChange("platform_credential")
		oPPRaw, nPPRaw := d.GetChange("platform_principal")

		if len(attributes) == 0 && isChangeSha256Removal(oPCRaw, nPCRaw) && isChangeSha256Removal(oPPRaw, nPPRaw) {
			return nil
		}

		attributes[PlatformApplicationAttributeNamePlatformCredential] = d.Get("platform_credential").(string)
		// If the platform requires a principal it must also be specified, even if it didn't change
		// since credential is stored as a hash, the only way to update principal is to update both
		// as they must be specified together in the request.
		if v, ok := d.GetOk("platform_principal"); ok {
			attributes[PlatformApplicationAttributeNamePlatformPrincipal] = v.(string)
		}
	}

	// Make API call to update attributes
	input := &sns.SetPlatformApplicationAttributesInput{
		Attributes:             aws.StringMap(attributes),
		PlatformApplicationArn: aws.String(d.Id()),
	}

	_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.SetPlatformApplicationAttributesWithContext(ctx, input)
	}, sns.ErrCodeInvalidParameterException, "is not a valid role to allow SNS to write to Cloudwatch Logs")

	if err != nil {
		return diag.Errorf("updating SNS Platform Application (%s): %s", d.Id(), err)
	}

	return resourcePlatformApplicationRead(ctx, d, meta)
}

func resourcePlatformApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	log.Printf("[DEBUG] Deleting SNS Platform Application: %s", d.Id())
	_, err := conn.DeletePlatformApplicationWithContext(ctx, &sns.DeletePlatformApplicationInput{
		PlatformApplicationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting SNS Platform Application (%s): %s", d.Id(), err)
	}

	return nil
}

func DecodePlatformApplicationID(input string) (arnS, name, platform string, err error) {
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
