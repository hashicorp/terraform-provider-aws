// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_emr_studio_session_mapping", name="Studio Session Mapping")
func resourceStudioSessionMapping() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStudioSessionMappingCreate,
		ReadWithoutTimeout:   resourceStudioSessionMappingRead,
		UpdateWithoutTimeout: resourceStudioSessionMappingUpdate,
		DeleteWithoutTimeout: resourceStudioSessionMappingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"identity_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"identity_id", "identity_name"},
			},
			"identity_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"identity_id", "identity_name"},
			},
			"identity_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IdentityType](),
			},
			"session_policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"studio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStudioSessionMappingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	var idOrName string
	studioId := d.Get("studio_id").(string)
	identityType := d.Get("identity_type").(string)
	input := &emr.CreateStudioSessionMappingInput{
		IdentityType:     awstypes.IdentityType(identityType),
		SessionPolicyArn: aws.String(d.Get("session_policy_arn").(string)),
		StudioId:         aws.String(studioId),
	}

	if v, ok := d.GetOk("identity_id"); ok {
		input.IdentityId = aws.String(v.(string))
		idOrName = v.(string)
	}

	if v, ok := d.GetOk("identity_name"); ok {
		input.IdentityName = aws.String(v.(string))
		idOrName = v.(string)
	}

	_, err := conn.CreateStudioSessionMapping(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Studio Session Mapping: %s", err)
	}

	d.SetId(studioSessionMappingCreateResourceID(studioId, identityType, idOrName))

	return append(diags, resourceStudioSessionMappingRead(ctx, d, meta)...)
}

func resourceStudioSessionMappingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	mapping, err := findStudioSessionMappingByIDOrName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Studio Session Mapping (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	d.Set("identity_id", mapping.IdentityId)
	d.Set("identity_name", mapping.IdentityName)
	d.Set("identity_type", mapping.IdentityType)
	d.Set("session_policy_arn", mapping.SessionPolicyArn)
	d.Set("studio_id", mapping.StudioId)

	return diags
}

func resourceStudioSessionMappingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	studioID, identityType, identityIDOrName, err := studioSessionMappingParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &emr.UpdateStudioSessionMappingInput{
		SessionPolicyArn: aws.String(d.Get("session_policy_arn").(string)),
		IdentityType:     awstypes.IdentityType(identityType),
		StudioId:         aws.String(studioID),
	}

	if isIdentityID(identityIDOrName) {
		input.IdentityId = aws.String(identityIDOrName)
	} else {
		input.IdentityName = aws.String(identityIDOrName)
	}

	_, err = conn.UpdateStudioSessionMapping(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	return append(diags, resourceStudioSessionMappingRead(ctx, d, meta)...)
}

func resourceStudioSessionMappingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	studioID, identityType, identityIDOrName, err := studioSessionMappingParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &emr.DeleteStudioSessionMappingInput{
		IdentityType: awstypes.IdentityType(identityType),
		StudioId:     aws.String(studioID),
	}

	if isIdentityID(identityIDOrName) {
		input.IdentityId = aws.String(identityIDOrName)
	} else {
		input.IdentityName = aws.String(identityIDOrName)
	}

	log.Printf("[INFO] Deleting EMR Studio Session Mapping: %s", d.Id())
	_, err = conn.DeleteStudioSessionMapping(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Studio session mapping does not exist.") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	return diags
}

const identityIDPattern = `([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}`

var identityIDPatternRegexp = regexache.MustCompile(identityIDPattern)

func isIdentityID(identityIdOrName string) bool {
	return identityIDPatternRegexp.MatchString(identityIdOrName)
}

const studioSessionMappingResourceIDSeparator = ":"

func studioSessionMappingCreateResourceID(studioID, identityType, identityIDOrName string) string {
	parts := []string{studioID, identityType, identityIDOrName}
	id := strings.Join(parts, studioSessionMappingResourceIDSeparator)

	return id
}

func studioSessionMappingParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, studioSessionMappingResourceIDSeparator)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected studio-id%[2]sidentity-type%[2]sidentity-id-or-name", id, studioSessionMappingResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findStudioSessionMappingByIDOrName(ctx context.Context, conn *emr.Client, id string) (*awstypes.SessionMappingDetail, error) {
	studioID, identityType, identityIDOrName, err := studioSessionMappingParseResourceID(id)
	if err != nil {
		return nil, err
	}

	input := &emr.GetStudioSessionMappingInput{
		StudioId:     aws.String(studioID),
		IdentityType: awstypes.IdentityType(identityType),
	}

	if isIdentityID(identityIDOrName) {
		input.IdentityId = aws.String(identityIDOrName)
	} else {
		input.IdentityName = aws.String(identityIDOrName)
	}

	output, err := conn.GetStudioSessionMapping(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Studio session mapping does not exist") ||
		errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Studio does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SessionMapping == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SessionMapping, nil
}
