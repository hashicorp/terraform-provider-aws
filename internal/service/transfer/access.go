// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_access", name="Access")
func resourceAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessCreate,
		ReadWithoutTimeout:   resourceAccessRead,
		UpdateWithoutTimeout: resourceAccessUpdate,
		DeleteWithoutTimeout: resourceAccessDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrExternalID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"home_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"home_directory_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						names.AttrTarget: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},
			"home_directory_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.HomeDirectoryTypePath,
				ValidateDiagFunc: enum.Validate[awstypes.HomeDirectoryType](),
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          verify.ValidIAMPolicyJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"posix_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Optional: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			names.AttrRole: {
				Type: schema.TypeString,
				// Although Role is required in the API it is not currently returned on Read.
				// Required:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validServerID,
			},
		},
	}
}

func resourceAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	externalID := d.Get(names.AttrExternalID).(string)
	serverID := d.Get("server_id").(string)
	id := accessCreateResourceID(serverID, externalID)
	input := &transfer.CreateAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	if v, ok := d.GetOk("home_directory"); ok {
		input.HomeDirectory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("home_directory_mappings"); ok {
		input.HomeDirectoryMappings = expandHomeDirectoryMapEntries(v.([]interface{}))
	}

	if v, ok := d.GetOk("home_directory_type"); ok {
		input.HomeDirectoryType = awstypes.HomeDirectoryType(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Policy = aws.String(policy)
	}

	if v, ok := d.GetOk("posix_profile"); ok {
		input.PosixProfile = expandPOSIXProfile(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrRole); ok {
		input.Role = aws.String(v.(string))
	}

	_, err := conn.CreateAccess(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Access (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAccessRead(ctx, d, meta)...)
}

func resourceAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, externalID, err := accessParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	access, err := findAccessByTwoPartKey(ctx, conn, serverID, externalID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Access (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Access (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrExternalID, access.ExternalId)
	d.Set("home_directory", access.HomeDirectory)
	if err := d.Set("home_directory_mappings", flattenHomeDirectoryMapEntries(access.HomeDirectoryMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting home_directory_mappings: %s", err)
	}
	d.Set("home_directory_type", access.HomeDirectoryType)
	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(access.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set(names.AttrPolicy, policyToSet)
	if err := d.Set("posix_profile", flattenPOSIXProfile(access.PosixProfile)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting posix_profile: %s", err)
	}
	// Role is currently not returned via the API.
	// d.Set("role", access.Role)
	d.Set("server_id", serverID)

	return diags
}

func resourceAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, externalID, err := accessParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &transfer.UpdateAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	if d.HasChange("home_directory") {
		input.HomeDirectory = aws.String(d.Get("home_directory").(string))
	}

	if d.HasChange("home_directory_mappings") {
		input.HomeDirectoryMappings = expandHomeDirectoryMapEntries(d.Get("home_directory_mappings").([]interface{}))
	}

	if d.HasChange("home_directory_type") {
		input.HomeDirectoryType = awstypes.HomeDirectoryType(d.Get("home_directory_type").(string))
	}

	if d.HasChange(names.AttrPolicy) {
		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Policy = aws.String(policy)
	}

	if d.HasChange("posix_profile") {
		input.PosixProfile = expandPOSIXProfile(d.Get("posix_profile").([]interface{}))
	}

	if d.HasChange(names.AttrRole) {
		input.Role = aws.String(d.Get(names.AttrRole).(string))
	}

	_, err = conn.UpdateAccess(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Transfer Access (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccessRead(ctx, d, meta)...)
}

func resourceAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, externalID, err := accessParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Transfer Access: %s", d.Id())
	_, err = conn.DeleteAccess(ctx, &transfer.DeleteAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Access (%s): %s", d.Id(), err)
	}

	return diags
}

const accessResourceIDSeparator = "/"

func accessCreateResourceID(serverID, externalID string) string {
	parts := []string{serverID, externalID}
	id := strings.Join(parts, accessResourceIDSeparator)

	return id
}

func accessParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, accessResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SERVERID%[2]sEXTERNALID", id, accessResourceIDSeparator)
}

func findAccessByTwoPartKey(ctx context.Context, conn *transfer.Client, serverID, externalID string) (*awstypes.DescribedAccess, error) {
	input := &transfer.DescribeAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	output, err := conn.DescribeAccess(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Access == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Access, nil
}
