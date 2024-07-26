// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_user")
func ResourceUser() *schema.Resource {
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
			"authentication_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthenticationType](),
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"first_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"last_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"send_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrUserName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName := d.Get(names.AttrUserName).(string)
	authType := d.Get("authentication_type").(string)

	input := &appstream.CreateUserInput{
		AuthenticationType: awstypes.AuthenticationType(authType),
		UserName:           aws.String(userName),
	}

	if v, ok := d.GetOk("first_name"); ok {
		input.FirstName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("last_name"); ok {
		input.LastName = aws.String(v.(string))
	}

	if !d.Get("send_email_notification").(bool) {
		input.MessageAction = awstypes.MessageActionSuppress
	}

	_, err := conn.CreateUser(ctx, input)

	id := EncodeUserID(userName, authType)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream User (%s): %s", id, err)
	}

	if _, err = waitUserAvailable(ctx, conn, userName, authType); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream User (%s) to be available: %s", id, err)
	}

	// Enabling/disabling workflow
	if !d.Get(names.AttrEnabled).(bool) {
		input := &appstream.DisableUserInput{
			AuthenticationType: awstypes.AuthenticationType(authType),
			UserName:           aws.String(userName),
		}

		_, err = conn.DisableUser(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling AppStream User (%s): %s", id, err)
		}
	}

	d.SetId(id)

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, err := DecodeUserID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding AppStream User ID (%s): %s", d.Id(), err)
	}

	user, err := FindUserByTwoPartKey(ctx, conn, userName, authType)
	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] AppStream User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream User (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, user.Arn)
	d.Set("authentication_type", user.AuthenticationType)
	d.Set(names.AttrCreatedTime, aws.ToTime(user.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrEnabled, user.Enabled)
	d.Set("first_name", user.FirstName)

	d.Set("last_name", user.LastName)
	d.Set(names.AttrUserName, user.UserName)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, err := DecodeUserID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding AppStream User ID (%s): %s", d.Id(), err)
	}

	if d.HasChange(names.AttrEnabled) {
		if d.Get(names.AttrEnabled).(bool) {
			input := &appstream.EnableUserInput{
				AuthenticationType: awstypes.AuthenticationType(authType),
				UserName:           aws.String(userName),
			}

			_, err = conn.EnableUser(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling AppStream User (%s): %s", d.Id(), err)
			}
		} else {
			input := &appstream.DisableUserInput{
				AuthenticationType: awstypes.AuthenticationType(authType),
				UserName:           aws.String(userName),
			}

			_, err = conn.DisableUser(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling AppStream User (%s): %s", d.Id(), err)
			}
		}
	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, err := DecodeUserID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding AppStream User ID (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteUser(ctx, &appstream.DeleteUserInput{
		AuthenticationType: awstypes.AuthenticationType(authType),
		UserName:           aws.String(userName),
	})

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting AppStream User (%s): %s", d.Id(), err)
	}

	return diags
}

func EncodeUserID(userName, authType string) string {
	return fmt.Sprintf("%s/%s", userName, authType)
}

func DecodeUserID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format UserName/AuthenticationType, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
