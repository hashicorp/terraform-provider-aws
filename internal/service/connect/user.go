// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_user", name="User")
// @Tags(identifierAttribute="arn")
func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
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
						"email": {
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
					},
				},
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"password": {
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
								if v := d.Get("phone_config.0.phone_type").(string); v == connect.PhoneTypeDeskPhone {
									return false
								}
								return true
							},
						},
						"phone_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(connect.PhoneType_Values(), false),
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

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)
	input := &connect.CreateUserInput{
		InstanceId:         aws.String(instanceID),
		PhoneConfig:        expandPhoneConfig(d.Get("phone_config").([]interface{})),
		RoutingProfileId:   aws.String(d.Get("routing_profile_id").(string)),
		SecurityProfileIds: flex.ExpandStringSet(d.Get("security_profile_ids").(*schema.Set)),
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
		input.IdentityInfo = expandIdentityInfo(v.([]interface{}))
	}

	if v, ok := d.GetOk("password"); ok {
		input.Password = aws.String(v.(string))
	}

	output, err := conn.CreateUserWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Connect User (%s): %s", name, err)
	}

	if output == nil {
		return diag.Errorf("creating Connect User (%s): empty output", name)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.UserId)))

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, userID, err := UserParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeUserWithContext(ctx, &connect.DescribeUserInput{
		InstanceId: aws.String(instanceID),
		UserId:     aws.String(userID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Connect User (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.User == nil {
		return diag.Errorf("getting Connect User (%s): empty response", d.Id())
	}

	user := resp.User

	d.Set("arn", user.Arn)
	d.Set("directory_user_id", user.DirectoryUserId)
	d.Set("hierarchy_group_id", user.HierarchyGroupId)
	d.Set("instance_id", instanceID)
	d.Set("name", user.Username)
	d.Set("routing_profile_id", user.RoutingProfileId)
	d.Set("security_profile_ids", flex.FlattenStringSet(user.SecurityProfileIds))
	d.Set("user_id", user.Id)

	if err := d.Set("identity_info", flattenIdentityInfo(user.IdentityInfo)); err != nil {
		return diag.Errorf("setting identity_info: %s", err)
	}

	if err := d.Set("phone_config", flattenPhoneConfig(user.PhoneConfig)); err != nil {
		return diag.Errorf("setting phone_config: %s", err)
	}

	setTagsOut(ctx, resp.User.Tags)

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, userID, err := UserParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
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

		_, err = conn.UpdateUserHierarchyWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating User hierarchy_group_id (%s): %s", d.Id(), err)
		}
	}

	// updates to identity_info
	if d.HasChange("identity_info") {
		input := &connect.UpdateUserIdentityInfoInput{
			IdentityInfo: expandIdentityInfo(d.Get("identity_info").([]interface{})),
			InstanceId:   aws.String(instanceID),
			UserId:       aws.String(userID),
		}

		_, err = conn.UpdateUserIdentityInfoWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating User identity_info (%s): %s", d.Id(), err)
		}
	}

	// updates to phone_config
	if d.HasChange("phone_config") {
		input := &connect.UpdateUserPhoneConfigInput{
			InstanceId:  aws.String(instanceID),
			PhoneConfig: expandPhoneConfig(d.Get("phone_config").([]interface{})),
			UserId:      aws.String(userID),
		}

		_, err = conn.UpdateUserPhoneConfigWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating User phone_config (%s): %s", d.Id(), err)
		}
	}

	// updates to routing_profile_id
	if d.HasChange("routing_profile_id") {
		input := &connect.UpdateUserRoutingProfileInput{
			InstanceId:       aws.String(instanceID),
			RoutingProfileId: aws.String(d.Get("routing_profile_id").(string)),
			UserId:           aws.String(userID),
		}

		_, err = conn.UpdateUserRoutingProfileWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating User routing_profile_id (%s): %s", d.Id(), err)
		}
	}

	// updates to security_profile_ids
	if d.HasChange("security_profile_ids") {
		input := &connect.UpdateUserSecurityProfilesInput{
			InstanceId:         aws.String(instanceID),
			SecurityProfileIds: flex.ExpandStringSet(d.Get("security_profile_ids").(*schema.Set)),
			UserId:             aws.String(userID),
		}

		_, err = conn.UpdateUserSecurityProfilesWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating User security_profile_ids (%s): %s", d.Id(), err)
		}
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, userID, err := UserParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteUserWithContext(ctx, &connect.DeleteUserInput{
		InstanceId: aws.String(instanceID),
		UserId:     aws.String(userID),
	})

	if err != nil {
		return diag.Errorf("deleting User (%s): %s", d.Id(), err)
	}

	return nil
}

func UserParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:userID", id)
	}

	return parts[0], parts[1], nil
}

func expandIdentityInfo(identityInfo []interface{}) *connect.UserIdentityInfo {
	if len(identityInfo) == 0 || identityInfo[0] == nil {
		return nil
	}

	tfMap, ok := identityInfo[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.UserIdentityInfo{}

	if v, ok := tfMap["email"].(string); ok && v != "" {
		result.Email = aws.String(v)
	}

	if v, ok := tfMap["first_name"].(string); ok && v != "" {
		result.FirstName = aws.String(v)
	}

	if v, ok := tfMap["last_name"].(string); ok && v != "" {
		result.LastName = aws.String(v)
	}

	return result
}

func expandPhoneConfig(phoneConfig []interface{}) *connect.UserPhoneConfig {
	if len(phoneConfig) == 0 || phoneConfig[0] == nil {
		return nil
	}

	tfMap, ok := phoneConfig[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.UserPhoneConfig{
		PhoneType: aws.String(tfMap["phone_type"].(string)),
	}

	if v, ok := tfMap["after_contact_work_time_limit"].(int); ok && v >= 0 {
		result.AfterContactWorkTimeLimit = aws.Int64(int64(v))
	}

	if v, ok := tfMap["auto_accept"].(bool); ok {
		result.AutoAccept = aws.Bool(v)
	}

	if v, ok := tfMap["desk_phone_number"].(string); ok && v != "" {
		result.DeskPhoneNumber = aws.String(v)
	}

	return result
}

func flattenIdentityInfo(identityInfo *connect.UserIdentityInfo) []interface{} {
	if identityInfo == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := identityInfo.Email; v != nil {
		values["email"] = aws.StringValue(v)
	}

	if v := identityInfo.FirstName; v != nil {
		values["first_name"] = aws.StringValue(v)
	}

	if v := identityInfo.LastName; v != nil {
		values["last_name"] = aws.StringValue(v)
	}

	return []interface{}{values}
}

func flattenPhoneConfig(phoneConfig *connect.UserPhoneConfig) []interface{} {
	if phoneConfig == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"phone_type": aws.StringValue(phoneConfig.PhoneType),
	}

	if v := phoneConfig.AfterContactWorkTimeLimit; v != nil {
		values["after_contact_work_time_limit"] = aws.Int64Value(v)
	}

	if v := phoneConfig.AutoAccept; v != nil {
		values["auto_accept"] = aws.BoolValue(v)
	}

	if v := phoneConfig.DeskPhoneNumber; v != nil {
		values["desk_phone_number"] = aws.StringValue(v)
	}

	return []interface{}{values}
}
