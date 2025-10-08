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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_user_stack_association", name="User Stack Association")
func resourceUserStackAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserStackAssociationCreate,
		ReadWithoutTimeout:   resourceUserStackAssociationRead,
		UpdateWithoutTimeout: schema.NoopContext, // TODO: Make send_email_notification ForceNew.
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

func resourceUserStackAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, stackName := d.Get(names.AttrUserName).(string), awstypes.AuthenticationType(d.Get("authentication_type").(string)), d.Get("stack_name").(string)
	id := userStackAssociationCreateResourceID(userName, authType, stackName)
	userStackAssociation := awstypes.UserStackAssociation{
		AuthenticationType: authType,
		StackName:          aws.String(stackName),
		UserName:           aws.String(userName),
	}

	if v, ok := d.GetOk("send_email_notification"); ok {
		userStackAssociation.SendEmailNotification = aws.Bool(v.(bool))
	}

	input := appstream.BatchAssociateUserStackInput{
		UserStackAssociations: []awstypes.UserStackAssociation{userStackAssociation},
	}
	output, err := conn.BatchAssociateUserStack(ctx, &input)

	if err == nil {
		err = userStackAssociationsError(output.Errors)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream User Stack Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceUserStackAssociationRead(ctx, d, meta)...)
}

func resourceUserStackAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, stackName, err := userStackAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	association, err := findUserStackAssociationByThreePartKey(ctx, conn, userName, authType, stackName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream User Stack Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream User Stack Association (%s): %s", d.Id(), err)
	}

	d.Set("authentication_type", association.AuthenticationType)
	d.Set("stack_name", association.StackName)
	d.Set(names.AttrUserName, association.UserName)

	return diags
}

func resourceUserStackAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, stackName, err := userStackAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting AppStream User Stack Association: %s", d.Id())
	input := appstream.BatchDisassociateUserStackInput{
		UserStackAssociations: []awstypes.UserStackAssociation{{
			AuthenticationType: authType,
			StackName:          aws.String(stackName),
			UserName:           aws.String(userName),
		}},
	}
	output, err := conn.BatchDisassociateUserStack(ctx, &input)

	if err == nil {
		err = userStackAssociationsError(output.Errors)
	}

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream User Stack Association (%s): %s", d.Id(), err)
	}

	return diags
}

const userStackAssociationResourceIDSeparator = "/"

func userStackAssociationCreateResourceID(userName string, authType awstypes.AuthenticationType, stackName string) string {
	parts := []string{userName, string(authType), stackName} // nosemgrep:ci.typed-enum-conversion
	id := strings.Join(parts, userStackAssociationResourceIDSeparator)

	return id
}

func userStackAssociationParseResourceID(id string) (string, awstypes.AuthenticationType, string, error) {
	parts := strings.SplitN(id, userStackAssociationResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected UserName%[2]sAuthenticationType%[2]sStackName", id, userStackAssociationResourceIDSeparator)
	}

	return parts[0], awstypes.AuthenticationType(parts[1]), parts[2], nil
}

func findUserStackAssociationByThreePartKey(ctx context.Context, conn *appstream.Client, userName string, authType awstypes.AuthenticationType, stackName string) (*awstypes.UserStackAssociation, error) {
	input := appstream.DescribeUserStackAssociationsInput{
		AuthenticationType: authType,
		StackName:          aws.String(stackName),
		UserName:           aws.String(userName),
	}

	return findUserStackAssociation(ctx, conn, &input)
}

func findUserStackAssociation(ctx context.Context, conn *appstream.Client, input *appstream.DescribeUserStackAssociationsInput) (*awstypes.UserStackAssociation, error) {
	output, err := findUserStackAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findUserStackAssociations(ctx context.Context, conn *appstream.Client, input *appstream.DescribeUserStackAssociationsInput) ([]awstypes.UserStackAssociation, error) {
	var output []awstypes.UserStackAssociation

	err := describeUserStackAssociationsPages(ctx, conn, input, func(page *appstream.DescribeUserStackAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.UserStackAssociations...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func userStackAssociationError(apiObject *awstypes.UserStackAssociationError) error {
	if apiObject == nil {
		return nil
	}

	return errs.APIError(apiObject.ErrorCode, aws.ToString(apiObject.ErrorMessage))
}

func userStackAssociationsError(apiObjects []awstypes.UserStackAssociationError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := userStackAssociationError(&apiObject); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
