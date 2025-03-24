// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	defaultGroupNamespace = "default"
)

// @SDKResource("aws_quicksight_group", name="Group")
func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrGroupName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrNamespace: {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  defaultGroupNamespace,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 63),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
					),
				},
			}
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	groupName := d.Get(names.AttrGroupName).(string)
	namespace := d.Get(names.AttrNamespace).(string)
	id := groupCreateResourceID(awsAccountID, namespace, groupName)
	input := &quicksight.CreateGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		Namespace:    aws.String(namespace),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Group (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, groupName, err := groupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	group, err := findGroupByThreePartKey(ctx, conn, awsAccountID, namespace, groupName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, group.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrGroupName, group.GroupName)
	d.Set(names.AttrNamespace, namespace)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, groupName, err := groupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &quicksight.UpdateGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		Namespace:    aws.String(namespace),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err = conn.UpdateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating QuickSight Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, groupName, err := groupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Group: %s", d.Id())
	_, err = conn.DeleteGroup(ctx, &quicksight.DeleteGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		Namespace:    aws.String(namespace),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Group (%s): %s", d.Id(), err)
	}

	return diags
}

const groupResourceIDSeparator = "/"

func groupCreateResourceID(awsAccountID, namespace, groupName string) string {
	parts := []string{awsAccountID, namespace, groupName}
	id := strings.Join(parts, groupResourceIDSeparator)

	return id
}

func groupParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, groupResourceIDSeparator, 3)

	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sNAMESPACE%[2]sGROUP_NAME", id, groupResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findGroupByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace, groupName string) (*awstypes.Group, error) {
	input := &quicksight.DescribeGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		Namespace:    aws.String(namespace),
	}

	return findGroup(ctx, conn, input)
}

func findGroup(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeGroupInput) (*awstypes.Group, error) {
	output, err := conn.DescribeGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Group == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Group, nil
}
