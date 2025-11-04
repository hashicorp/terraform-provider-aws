// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_user", name="User")
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

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"hierarchy_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"identity_info": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEmail: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"first_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 100),
						},
						"last_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 100),
						},
						"secondary_email": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrPassword: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 64),
			},
			"phone_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_contact_work_time_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"auto_accept": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"desk_phone_number": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validDeskPhoneNumber,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := awstypes.PhoneType(d.Get("phone_config.0.phone_type").(string)); v == awstypes.PhoneTypeDeskPhone {
									return false
								}
								return true
							},
						},
						"phone_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PhoneType](),
						},
					},
				},
			},
			"routing_profile_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"security_profile_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	input := &connect.CreateUserInput{
		InstanceId:         aws.String(instanceID),
		PhoneConfig:        expandUserPhoneConfig(d.Get("phone_config").([]any)),
		RoutingProfileId:   aws.String(d.Get("routing_profile_id").(string)),
		SecurityProfileIds: flex.ExpandStringValueSet(d.Get("security_profile_ids").(*schema.Set)),
		Tags:               getTagsIn(ctx),
		Username:           aws.String(name),
	}

	if v, ok := d.GetOk("directory_user_id"); ok {
		input.DirectoryUserId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hierarchy_group_id"); ok {
		input.HierarchyGroupId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("identity_info"); ok {
		input.IdentityInfo = expandUserIdentityInfo(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrPassword); ok {
		input.Password = aws.String(v.(string))
	}

	output, err := conn.CreateUser(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect User (%s): %s", name, err)
	}

	id := userCreateResourceID(instanceID, aws.ToString(output.UserId))
	d.SetId(id)

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, userID, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	user, err := findUserByTwoPartKey(ctx, conn, instanceID, userID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect User (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, user.Arn)
	d.Set("directory_user_id", user.DirectoryUserId)
	d.Set("hierarchy_group_id", user.HierarchyGroupId)
	if err := d.Set("identity_info", flattenUserIdentityInfo(user.IdentityInfo)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting identity_info: %s", err)
	}
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, user.Username)
	if err := d.Set("phone_config", flattenUserPhoneConfig(user.PhoneConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting phone_config: %s", err)
	}
	d.Set("routing_profile_id", user.RoutingProfileId)
	d.Set("security_profile_ids", user.SecurityProfileIds)
	d.Set("user_id", user.Id)

	setTagsOut(ctx, user.Tags)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, userID, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// User has 5 update APIs
	// UpdateUserHierarchyWithContext: Assigns the specified hierarchy group to the specified user.
	// UpdateUserIdentityInfoWithContext: Updates the identity information for the specified user.
	// UpdateUserPhoneConfigWithContext: Updates the phone configuration settings for the specified user.
	// UpdateUserRoutingProfileWithContext: Assigns the specified routing profile to the specified user.
	// UpdateUserSecurityProfilesWithContext: Assigns the specified security profiles to the specified user.

	// updates to hierarchy_group_id
	if d.HasChange("hierarchy_group_id") {
		input := &connect.UpdateUserHierarchyInput{
			InstanceId: aws.String(instanceID),
			UserId:     aws.String(userID),
		}

		if v, ok := d.GetOk("hierarchy_group_id"); ok {
			input.HierarchyGroupId = aws.String(v.(string))
		}

		_, err = conn.UpdateUserHierarchy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect User (%s) HierarchyGroupId: %s", d.Id(), err)
		}
	}

	// updates to identity_info
	if d.HasChange("identity_info") {
		input := &connect.UpdateUserIdentityInfoInput{
			IdentityInfo: expandUserIdentityInfo(d.Get("identity_info").([]any)),
			InstanceId:   aws.String(instanceID),
			UserId:       aws.String(userID),
		}

		_, err = conn.UpdateUserIdentityInfo(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect User (%s) IdentityInfo: %s", d.Id(), err)
		}
	}

	// updates to phone_config
	if d.HasChange("phone_config") {
		input := &connect.UpdateUserPhoneConfigInput{
			InstanceId:  aws.String(instanceID),
			PhoneConfig: expandUserPhoneConfig(d.Get("phone_config").([]any)),
			UserId:      aws.String(userID),
		}

		_, err = conn.UpdateUserPhoneConfig(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect User (%s) PhoneConfig: %s", d.Id(), err)
		}
	}

	// updates to routing_profile_id
	if d.HasChange("routing_profile_id") {
		input := &connect.UpdateUserRoutingProfileInput{
			InstanceId:       aws.String(instanceID),
			RoutingProfileId: aws.String(d.Get("routing_profile_id").(string)),
			UserId:           aws.String(userID),
		}

		_, err = conn.UpdateUserRoutingProfile(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect User (%s) RoutingProfileId: %s", d.Id(), err)
		}
	}

	// updates to security_profile_ids
	if d.HasChange("security_profile_ids") {
		input := &connect.UpdateUserSecurityProfilesInput{
			InstanceId:         aws.String(instanceID),
			SecurityProfileIds: flex.ExpandStringValueSet(d.Get("security_profile_ids").(*schema.Set)),
			UserId:             aws.String(userID),
		}

		_, err = conn.UpdateUserSecurityProfiles(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect User (%s) SecurityProfileIds: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, userID, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect User: %s", d.Id())
	input := connect.DeleteUserInput{
		InstanceId: aws.String(instanceID),
		UserId:     aws.String(userID),
	}
	_, err = conn.DeleteUser(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect User (%s): %s", d.Id(), err)
	}

	return diags
}

const userResourceIDSeparator = ":"

func userCreateResourceID(instanceID, userID string) string {
	parts := []string{instanceID, userID}
	id := strings.Join(parts, userResourceIDSeparator)

	return id
}

func userParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, userResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]suserID", id, userResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findUserByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, userID string) (*awstypes.User, error) {
	input := &connect.DescribeUserInput{
		InstanceId: aws.String(instanceID),
		UserId:     aws.String(userID),
	}

	return findUser(ctx, conn, input)
}

func findUser(ctx context.Context, conn *connect.Client, input *connect.DescribeUserInput) (*awstypes.User, error) {
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

func expandUserIdentityInfo(tfList []any) *awstypes.UserIdentityInfo {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.UserIdentityInfo{}

	if v, ok := tfMap[names.AttrEmail].(string); ok && v != "" {
		apiObject.Email = aws.String(v)
	}

	if v, ok := tfMap["first_name"].(string); ok && v != "" {
		apiObject.FirstName = aws.String(v)
	}

	if v, ok := tfMap["last_name"].(string); ok && v != "" {
		apiObject.LastName = aws.String(v)
	}

	if v, ok := tfMap["secondary_email"].(string); ok && v != "" {
		apiObject.SecondaryEmail = aws.String(v)
	}

	return apiObject
}

func expandUserPhoneConfig(tfList []any) *awstypes.UserPhoneConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.UserPhoneConfig{
		PhoneType: awstypes.PhoneType(tfMap["phone_type"].(string)),
	}

	if v, ok := tfMap["after_contact_work_time_limit"].(int); ok && v >= 0 {
		apiObject.AfterContactWorkTimeLimit = int32(v)
	}

	if v, ok := tfMap["auto_accept"].(bool); ok {
		apiObject.AutoAccept = v
	}

	if v, ok := tfMap["desk_phone_number"].(string); ok && v != "" {
		apiObject.DeskPhoneNumber = aws.String(v)
	}

	return apiObject
}

func flattenUserIdentityInfo(apiObject *awstypes.UserIdentityInfo) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if v := apiObject.Email; v != nil {
		tfMap[names.AttrEmail] = aws.ToString(v)
	}

	if v := apiObject.FirstName; v != nil {
		tfMap["first_name"] = aws.ToString(v)
	}

	if v := apiObject.LastName; v != nil {
		tfMap["last_name"] = aws.ToString(v)
	}

	if v := apiObject.SecondaryEmail; v != nil {
		tfMap["secondary_email"] = aws.ToString(v)
	}

	return []any{tfMap}
}

func flattenUserPhoneConfig(apiObject *awstypes.UserPhoneConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"after_contact_work_time_limit": apiObject.AfterContactWorkTimeLimit,
		"auto_accept":                   apiObject.AutoAccept,
		"phone_type":                    apiObject.PhoneType,
	}

	if v := apiObject.DeskPhoneNumber; v != nil {
		tfMap["desk_phone_number"] = aws.ToString(v)
	}

	return []any{tfMap}
}
