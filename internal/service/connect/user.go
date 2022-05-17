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
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceUserRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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
		return diag.FromErr(fmt.Errorf("error getting Connect User (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.User == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect User (%s): empty response", d.Id()))
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
		return diag.FromErr(fmt.Errorf("error setting identity_info: %w", err))
	}

	if err := d.Set("phone_config", flattenPhoneConfig(user.PhoneConfig)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting phone_config: %w", err))
	}

	tags := KeyValueTags(resp.User.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
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
