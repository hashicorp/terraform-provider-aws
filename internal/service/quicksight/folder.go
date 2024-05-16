// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_folder", name="Folder")
// @Tags(identifierAttribute="arn")
func ResourceFolder() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"created_time": {
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      quicksight.FolderTypeShared,
				ValidateFunc: validation.StringInSlice(quicksight.FolderType_Values(), false),
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
			"permissions": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 64,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 16,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"principal": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameFolder = "Folder"
)

func resourceFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}

	folderId := d.Get("folder_id").(string)

	d.SetId(createFolderId(awsAccountId, folderId))

	in := &quicksight.CreateFolderInput{
		AwsAccountId: aws.String(awsAccountId),
		FolderId:     aws.String(folderId),
		Name:         aws.String(d.Get("name").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("folder_type"); ok {
		in.FolderType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parent_folder_arn"); ok {
		in.ParentFolderArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("permissions"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Permissions = expandResourcePermissions(v.([]interface{}))
	}

	out, err := conn.CreateFolderWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.QuickSight, create.ErrActionCreating, ResNameFolder, d.Get("name").(string), err)
	}

	if out == nil || out.Arn == nil {
		return create.DiagError(names.QuickSight, create.ErrActionCreating, ResNameFolder, d.Get("name").(string), errors.New("empty output"))
	}

	return resourceFolderRead(ctx, d, meta)
}

func resourceFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, folderId, err := ParseFolderId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	out, err := FindFolderByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Folder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.QuickSight, create.ErrActionReading, ResNameFolder, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("aws_account_id", awsAccountId)
	d.Set("created_time", out.CreatedTime.Format(time.RFC3339))
	d.Set("folder_id", out.FolderId)
	d.Set("folder_type", out.FolderType)
	d.Set("last_updated_time", out.LastUpdatedTime.Format(time.RFC3339))
	d.Set("name", out.Name)

	if len(out.FolderPath) > 0 {
		d.Set("parent_folder_arn", out.FolderPath[len(out.FolderPath)-1])
	}

	if err := d.Set("folder_path", flex.FlattenStringList(out.FolderPath)); err != nil {
		return diag.Errorf("setting folder_path: %s", err)
	}

	permsResp, err := conn.DescribeFolderPermissionsWithContext(ctx, &quicksight.DescribeFolderPermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		FolderId:     aws.String(folderId),
	})

	if err != nil {
		return diag.Errorf("describing QuickSight Folder (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set("permissions", flattenPermissions(permsResp.Permissions)); err != nil {
		return diag.Errorf("setting permissions: %s", err)
	}
	return nil
}

func resourceFolderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, folderId, err := ParseFolderId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChangesExcept("permission", "tags", "tags_all") {
		in := &quicksight.UpdateFolderInput{
			AwsAccountId: aws.String(awsAccountId),
			FolderId:     aws.String(folderId),
			Name:         aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating QuickSight Folder (%s): %#v", d.Id(), in)
		_, err = conn.UpdateFolderWithContext(ctx, in)
		if err != nil {
			return create.DiagError(names.QuickSight, create.ErrActionUpdating, ResNameFolder, d.Id(), err)
		}
	}

	if d.HasChange("permissions") {
		oraw, nraw := d.GetChange("permissions")
		o := oraw.([]interface{})
		n := nraw.([]interface{})

		toGrant, toRevoke := DiffPermissions(o, n)

		params := &quicksight.UpdateFolderPermissionsInput{
			AwsAccountId: aws.String(awsAccountId),
			FolderId:     aws.String(folderId),
		}

		if len(toGrant) > 0 {
			params.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			params.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateFolderPermissionsWithContext(ctx, params)

		if err != nil {
			return diag.Errorf("updating QuickSight Folder (%s) permissions: %s", folderId, err)
		}
	}

	return resourceFolderRead(ctx, d, meta)
}

func resourceFolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	log.Printf("[INFO] Deleting QuickSight Folder %s", d.Id())

	awsAccountId, folderId, err := ParseFolderId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteFolderWithContext(ctx, &quicksight.DeleteFolderInput{
		AwsAccountId: aws.String(awsAccountId),
		FolderId:     aws.String(folderId),
	})

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.QuickSight, create.ErrActionDeleting, ResNameFolder, d.Id(), err)
	}

	return nil
}

func FindFolderByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.Folder, error) {
	awsAccountId, folderId, err := ParseFolderId(id)
	if err != nil {
		return nil, err
	}

	descOpts := &quicksight.DescribeFolderInput{
		AwsAccountId: aws.String(awsAccountId),
		FolderId:     aws.String(folderId),
	}

	out, err := conn.DescribeFolderWithContext(ctx, descOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: descOpts,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Folder == nil {
		return nil, tfresource.NewEmptyResultError(descOpts)
	}

	return out.Folder, nil
}

func ParseFolderId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,FOLDER_ID", id)
	}
	return parts[0], parts[1], nil
}

func createFolderId(awsAccountID, folderId string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, folderId)
}
