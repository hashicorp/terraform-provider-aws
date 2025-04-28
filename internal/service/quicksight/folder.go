// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_folder", name="Folder")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.Folder")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func resourceFolder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFolderCreate,
		ReadWithoutTimeout:   resourceFolderRead,
		UpdateWithoutTimeout: resourceFolderUpdate,
		DeleteWithoutTimeout: resourceFolderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAWSAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.NoZeroValues,
					validation.StringLenBetween(1, 2048),
				),
			},
			"folder_path": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"folder_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.FolderTypeShared,
				ValidateDiagFunc: enum.Validate[awstypes.FolderType](),
			},
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.NoZeroValues,
					validation.StringLenBetween(1, 200),
				),
			},
			"parent_folder_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPermissions: quicksightschema.PermissionsSchema(),
			names.AttrTags:        tftags.TagsSchema(),
			names.AttrTagsAll:     tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFolderCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	folderID := d.Get("folder_id").(string)
	id := folderCreateResourceID(awsAccountID, folderID)
	input := &quicksight.CreateFolderInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("folder_type"); ok {
		input.FolderType = awstypes.FolderType(v.(string))
	}

	if v, ok := d.GetOk("parent_folder_arn"); ok {
		input.ParentFolderArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	_, err := conn.CreateFolder(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Folder (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceFolderRead(ctx, d, meta)...)
}

func resourceFolderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, folderID, err := folderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	folder, err := findFolderByTwoPartKey(ctx, conn, awsAccountID, folderID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Folder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Folder (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, folder.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrCreatedTime, folder.CreatedTime.Format(time.RFC3339))
	d.Set("folder_id", folder.FolderId)
	d.Set("folder_type", folder.FolderType)
	d.Set(names.AttrLastUpdatedTime, folder.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, folder.Name)
	if len(folder.FolderPath) > 0 {
		d.Set("parent_folder_arn", folder.FolderPath[len(folder.FolderPath)-1])
	}
	d.Set("folder_path", folder.FolderPath)

	permissions, err := findFolderPermissionsByTwoPartKey(ctx, conn, awsAccountID, folderID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Folder (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceFolderUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, folderID, err := folderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept("permission", names.AttrTags, names.AttrTagsAll) {
		input := &quicksight.UpdateFolderInput{
			AwsAccountId: aws.String(awsAccountID),
			FolderId:     aws.String(folderID),
			Name:         aws.String(d.Get(names.AttrName).(string)),
		}

		_, err = conn.UpdateFolder(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Folder (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		o, n := d.GetChange(names.AttrPermissions)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateFolderPermissionsInput{
			AwsAccountId: aws.String(awsAccountID),
			FolderId:     aws.String(folderID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateFolderPermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Folder (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFolderRead(ctx, d, meta)...)
}

func resourceFolderDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, folderID, err := folderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Folder: %s", d.Id())
	_, err = conn.DeleteFolder(ctx, &quicksight.DeleteFolderInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Folder (%s): %s", d.Id(), err)
	}

	return diags
}

const folderResourceIDSeparator = ","

func folderCreateResourceID(awsAccountID, folderID string) string {
	parts := []string{awsAccountID, folderID}
	id := strings.Join(parts, folderResourceIDSeparator)

	return id
}

func folderParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, folderResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sFOLDER_ID", id, folderResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findFolderByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, folderID string) (*awstypes.Folder, error) {
	input := &quicksight.DescribeFolderInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
	}

	return findFolder(ctx, conn, input)
}

func findFolder(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeFolderInput) (*awstypes.Folder, error) {
	output, err := conn.DescribeFolder(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Folder == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Folder, nil
}

func findFolderPermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, folderID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeFolderPermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
	}

	return findFolderPermissions(ctx, conn, input)
}

func findFolderPermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeFolderPermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeFolderPermissions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Permissions, nil
}
