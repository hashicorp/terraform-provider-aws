// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(appstream.AuthenticationType_Values(), false),
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
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
			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn(ctx)

	userName := d.Get("user_name").(string)
	authType := d.Get("authentication_type").(string)

	input := &appstream.CreateUserInput{
		AuthenticationType: aws.String(authType),
		UserName:           aws.String(userName),
	}

	if v, ok := d.GetOk("first_name"); ok {
		input.FirstName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("last_name"); ok {
		input.LastName = aws.String(v.(string))
	}

	if !d.Get("send_email_notification").(bool) {
		input.MessageAction = aws.String(appstream.MessageActionSuppress)
	}

	_, err := conn.CreateUserWithContext(ctx, input)

	id := EncodeUserID(userName, authType)

	if err != nil {
		return diag.Errorf("creating AppStream User (%s): %s", id, err)
	}

	if _, err = waitUserAvailable(ctx, conn, userName, authType); err != nil {
		return diag.Errorf("waiting for AppStream User (%s) to be available: %s", id, err)
	}

	// Enabling/disabling workflow
	if !d.Get("enabled").(bool) {
		input := &appstream.DisableUserInput{
			AuthenticationType: aws.String(authType),
			UserName:           aws.String(userName),
		}

		_, err = conn.DisableUserWithContext(ctx, input)
		if err != nil {
			return diag.Errorf("disabling AppStream User (%s): %s", id, err)
		}
	}

	d.SetId(id)

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn(ctx)

	userName, authType, err := DecodeUserID(d.Id())
	if err != nil {
		return diag.Errorf("decoding AppStream User ID (%s): %s", d.Id(), err)
	}

	user, err := FindUserByUserNameAndAuthType(ctx, conn, userName, authType)
	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] AppStream User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("reading AppStream User (%s): %s", d.Id(), err)
	}

	d.Set("arn", user.Arn)
	d.Set("authentication_type", user.AuthenticationType)
	d.Set("created_time", aws.TimeValue(user.CreatedTime).Format(time.RFC3339))
	d.Set("enabled", user.Enabled)
	d.Set("first_name", user.FirstName)

	d.Set("last_name", user.LastName)
	d.Set("user_name", user.UserName)

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn(ctx)

	userName, authType, err := DecodeUserID(d.Id())
	if err != nil {
		return diag.Errorf("decoding AppStream User ID (%s): %s", d.Id(), err)
	}

	if d.HasChange("enabled") {
		if d.Get("enabled").(bool) {
			input := &appstream.EnableUserInput{
				AuthenticationType: aws.String(authType),
				UserName:           aws.String(userName),
			}

			_, err = conn.EnableUserWithContext(ctx, input)
			if err != nil {
				return diag.Errorf("enabling AppStream User (%s): %s", d.Id(), err)
			}
		} else {
			input := &appstream.DisableUserInput{
				AuthenticationType: aws.String(authType),
				UserName:           aws.String(userName),
			}

			_, err = conn.DisableUserWithContext(ctx, input)
			if err != nil {
				return diag.Errorf("disabling AppStream User (%s): %s", d.Id(), err)
			}
		}
	}

	return nil
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn(ctx)

	userName, authType, err := DecodeUserID(d.Id())
	if err != nil {
		return diag.Errorf("decoding AppStream User ID (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteUserWithContext(ctx, &appstream.DeleteUserInput{
		AuthenticationType: aws.String(authType),
		UserName:           aws.String(userName),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("deleting AppStream User (%s): %s", d.Id(), err)
	}

	return nil
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
