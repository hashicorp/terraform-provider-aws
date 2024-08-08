// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// @SDKResource("aws_memorydb_user", name="User")
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
			"access_string": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_mode": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"passwords": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							MaxItems: 2,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(16, 128),
							},
							Set:       schema.HashString,
							Sensitive: true,
						},
						"password_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.InputAuthenticationType](),
						},
					},
				},
			},
			"minimum_engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrUserName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateResourceName(userNameMaxLength),
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	userName := d.Get(names.AttrUserName).(string)
	input := &memorydb.CreateUserInput{
		AccessString: aws.String(d.Get("access_string").(string)),
		Tags:         getTagsIn(ctx),
		UserName:     aws.String(userName),
	}

	if v, ok := d.GetOk("authentication_mode"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AuthenticationMode = expandAuthenticationMode(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateUser(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB User (%s): %s", userName, err)
	}

	d.SetId(userName)

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	user, err := FindUserByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MemoryDB User (%s): %s", d.Id(), err)
	}

	d.Set("access_string", user.AccessString)
	d.Set(names.AttrARN, user.ARN)

	if v := user.Authentication; v != nil {
		authenticationMode := map[string]interface{}{
			"passwords":      d.Get("authentication_mode.0.passwords"),
			"password_count": aws.ToInt32(v.PasswordCount),
			names.AttrType:   string(v.Type),
		}

		if err := d.Set("authentication_mode", []interface{}{authenticationMode}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting authentication_mode: %s", err)
		}
	}

	d.Set("minimum_engine_version", user.MinimumEngineVersion)
	d.Set(names.AttrUserName, user.Name)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &memorydb.UpdateUserInput{
			UserName: aws.String(d.Id()),
		}

		if d.HasChange("access_string") {
			input.AccessString = aws.String(d.Get("access_string").(string))
		}

		if v, ok := d.GetOk("authentication_mode"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AuthenticationMode = expandAuthenticationMode(v.([]interface{})[0].(map[string]interface{}))
		}

		_, err := conn.UpdateUser(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MemoryDB User (%s): %s", d.Id(), err)
		}

		if err := waitUserActive(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB User (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB User: (%s)", d.Id())
	_, err := conn.DeleteUser(ctx, &memorydb.DeleteUserInput{
		UserName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.UserNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB User (%s): %s", d.Id(), err)
	}

	if err := waitUserDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB User (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandAuthenticationMode(tfMap map[string]interface{}) *awstypes.AuthenticationMode {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AuthenticationMode{}

	if v, ok := tfMap["passwords"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Passwords = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.InputAuthenticationType(v)
	}

	return apiObject
}
