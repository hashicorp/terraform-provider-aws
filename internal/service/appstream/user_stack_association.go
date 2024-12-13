// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_user_stack_association")
func ResourceUserStackAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserStackAssociationCreate,
		ReadWithoutTimeout:   resourceUserStackAssociationRead,
		UpdateWithoutTimeout: schema.NoopContext,
		DeleteWithoutTimeout: resourceUserStackAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"authentication_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthenticationType](),
			},
			"send_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"stack_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserStackAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	input := awstypes.UserStackAssociation{
		AuthenticationType: awstypes.AuthenticationType(d.Get("authentication_type").(string)),
		StackName:          aws.String(d.Get("stack_name").(string)),
		UserName:           aws.String(d.Get(names.AttrUserName).(string)),
	}

	if v, ok := d.GetOk("send_email_notification"); ok {
		input.SendEmailNotification = aws.Bool(v.(bool))
	}

	id := EncodeUserStackAssociationID(d.Get(names.AttrUserName).(string), d.Get("authentication_type").(string), d.Get("stack_name").(string))

	output, err := conn.BatchAssociateUserStack(ctx, &appstream.BatchAssociateUserStackInput{
		UserStackAssociations: []awstypes.UserStackAssociation{input},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream User Stack Association (%s): %s", id, err)
	}
	if len(output.Errors) > 0 {
		var errs []error

		for _, err := range output.Errors {
			errs = append(errs, fmt.Errorf("%s: %s", string(err.ErrorCode), aws.ToString(err.ErrorMessage)))
		}

		return sdkdiag.AppendErrorf(diags, "creating AppStream User Stack Association (%s): %s", id, errors.Join(errs...))
	}

	d.SetId(id)

	return append(diags, resourceUserStackAssociationRead(ctx, d, meta)...)
}

func resourceUserStackAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, stackName, err := DecodeUserStackAssociationID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding AppStream User Stack Association ID (%s): %s", d.Id(), err)
	}

	resp, err := conn.DescribeUserStackAssociations(ctx,
		&appstream.DescribeUserStackAssociationsInput{
			AuthenticationType: awstypes.AuthenticationType(authType),
			StackName:          aws.String(stackName),
			UserName:           aws.String(userName),
		})

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] AppStream User Stack Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream User Stack Association (%s): %s", d.Id(), err)
	}

	if resp == nil || len(resp.UserStackAssociations) == 0 {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading AppStream User Stack Association (%s): empty output after creation", d.Id())
		}
		log.Printf("[WARN] AppStream User Stack Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	association := resp.UserStackAssociations[0]

	d.Set("authentication_type", association.AuthenticationType)
	d.Set("stack_name", association.StackName)
	d.Set(names.AttrUserName, association.UserName)

	return diags
}

func resourceUserStackAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, stackName, err := DecodeUserStackAssociationID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding AppStream User Stack Association ID (%s): %s", d.Id(), err)
	}

	input := awstypes.UserStackAssociation{
		AuthenticationType: awstypes.AuthenticationType(authType),
		StackName:          aws.String(stackName),
		UserName:           aws.String(userName),
	}

	_, err = conn.BatchDisassociateUserStack(ctx, &appstream.BatchDisassociateUserStackInput{
		UserStackAssociations: []awstypes.UserStackAssociation{input},
	})

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting AppStream User Stack Association (%s): %s", d.Id(), err)
	}
	return diags
}

func EncodeUserStackAssociationID(userName, authType, stackName string) string {
	return fmt.Sprintf("%s/%s/%s", userName, authType, stackName)
}

func DecodeUserStackAssociationID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format UserName/AuthenticationType/StackName, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}
