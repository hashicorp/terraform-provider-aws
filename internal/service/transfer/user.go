// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_user", name="User")
// @Tags(identifierAttribute="arn")
func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"home_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"home_directory_mappings": {
				Type:     schema.TypeList,
				Optional: true,
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validServerID,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrUserName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserName,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID := d.Get("server_id").(string)
	userName := d.Get(names.AttrUserName).(string)
	id := userCreateResourceID(serverID, userName)
	input := &transfer.CreateUserInput{
		Role:     aws.String(d.Get(names.AttrRole).(string)),
		ServerId: aws.String(serverID),
		Tags:     getTagsIn(ctx),
		UserName: aws.String(userName),
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
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", v.(string), err)
		}

		input.Policy = aws.String(policy)
	}

	if v, ok := d.GetOk("posix_profile"); ok {
		input.PosixProfile = expandPOSIXProfile(v.([]interface{}))
	}

	_, err := conn.CreateUser(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer User (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, userName, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	user, err := findUserByTwoPartKey(ctx, conn, serverID, userName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer User (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, user.Arn)
	d.Set("home_directory", user.HomeDirectory)
	if err := d.Set("home_directory_mappings", flattenHomeDirectoryMapEntries(user.HomeDirectoryMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting home_directory_mappings: %s", err)
	}
	d.Set("home_directory_type", user.HomeDirectoryType)

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(user.Policy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer User (%s): %s", d.Id(), err)
	}
	d.Set(names.AttrPolicy, policyToSet)

	if err := d.Set("posix_profile", flattenPOSIXProfile(user.PosixProfile)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting posix_profile: %s", err)
	}
	d.Set(names.AttrRole, user.Role)
	d.Set("server_id", serverID)
	d.Set(names.AttrUserName, user.UserName)

	setTagsOut(ctx, user.Tags)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, userName, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &transfer.UpdateUserInput{
			ServerId: aws.String(serverID),
			UserName: aws.String(userName),
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
				return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get(names.AttrPolicy).(string), err)
			}

			input.Policy = aws.String(policy)
		}

		if d.HasChange("posix_profile") {
			input.PosixProfile = expandPOSIXProfile(d.Get("posix_profile").([]interface{}))
		}

		if d.HasChange(names.AttrRole) {
			input.Role = aws.String(d.Get(names.AttrRole).(string))
		}

		_, err = conn.UpdateUser(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer User (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, userName, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := userDelete(ctx, conn, serverID, userName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

const userResourceIDSeparator = "/"

func userCreateResourceID(serverID, userName string) string {
	parts := []string{serverID, userName}
	id := strings.Join(parts, userResourceIDSeparator)

	return id
}

func userParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, userResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SERVERID%[2]sUSERNAME", id, userResourceIDSeparator)
}

func findUserByTwoPartKey(ctx context.Context, conn *transfer.Client, serverID, userName string) (*awstypes.DescribedUser, error) {
	input := &transfer.DescribeUserInput{
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
	}

	output, err := conn.DescribeUser(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.User == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.User, nil
}

func userDelete(ctx context.Context, conn *transfer.Client, serverID, userName string, timeout time.Duration) error {
	id := userCreateResourceID(serverID, userName)
	input := &transfer.DeleteUserInput{
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
	}

	log.Printf("[INFO] Deleting Transfer User: %s", id)
	_, err := conn.DeleteUser(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Transfer User (%s): %w", id, err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, timeout, func() (interface{}, error) {
		return findUserByTwoPartKey(ctx, conn, serverID, userName)
	})

	if err != nil {
		return fmt.Errorf("waiting for Transfer User (%s) delete: %w", id, err)
	}

	return nil
}

func expandHomeDirectoryMapEntries(tfList []interface{}) []awstypes.HomeDirectoryMapEntry {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.HomeDirectoryMapEntry, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.HomeDirectoryMapEntry{
			Entry:  aws.String(tfMap["entry"].(string)),
			Target: aws.String(tfMap[names.AttrTarget].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenHomeDirectoryMapEntries(apiObjects []awstypes.HomeDirectoryMapEntry) []interface{} {
	tfList := make([]interface{}, len(apiObjects))

	for i, apiObject := range apiObjects {
		tfList[i] = map[string]interface{}{
			"entry":          aws.ToString(apiObject.Entry),
			names.AttrTarget: aws.ToString(apiObject.Target),
		}
	}

	return tfList
}

func expandPOSIXProfile(tfList []interface{}) *awstypes.PosixProfile {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.PosixProfile{
		Gid: aws.Int64(int64(tfMap["gid"].(int))),
		Uid: aws.Int64(int64(tfMap["uid"].(int))),
	}

	if v, ok := tfMap["secondary_gids"].(*schema.Set); ok && len(v.List()) > 0 {
		apiObject.SecondaryGids = flex.ExpandInt64ValueSet(v)
	}

	return apiObject
}

func flattenPOSIXProfile(apiObject *awstypes.PosixProfile) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"gid":            aws.ToInt64(apiObject.Gid),
		"secondary_gids": apiObject.SecondaryGids,
		"uid":            aws.ToInt64(apiObject.Uid),
	}

	return []interface{}{tfMap}
}
