// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_directory_service_shared_directory", name="Shared Directory")
func resourceSharedDirectory() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSharedDirectoryCreate,
		ReadWithoutTimeout:   resourceSharedDirectoryRead,
		DeleteWithoutTimeout: resourceSharedDirectoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"method": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.ShareMethodHandshake,
				ValidateDiagFunc: enum.Validate[awstypes.ShareMethod](),
			},
			"notes": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"shared_directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTarget: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.TargetTypeAccount,
							ValidateDiagFunc: enum.Validate[awstypes.TargetType](),
						},
					},
				},
			},
		},
	}
}

func resourceSharedDirectoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID := d.Get("directory_id").(string)
	input := &directoryservice.ShareDirectoryInput{
		DirectoryId: aws.String(directoryID),
		ShareMethod: awstypes.ShareMethod(d.Get("method").(string)),
		ShareTarget: expandShareTarget(d.Get(names.AttrTarget).([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk("notes"); ok {
		input.ShareNotes = aws.String(v.(string))
	}

	output, err := conn.ShareDirectory(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Directory Service Shared Directory (%s): %s", directoryID, err)
	}

	d.SetId(sharedDirectoryCreateResourceID(directoryID, aws.ToString(output.SharedDirectoryId)))

	return append(diags, resourceSharedDirectoryRead(ctx, d, meta)...)
}

func resourceSharedDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	ownerDirID, sharedDirID, err := sharedDirectoryParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findSharedDirectoryByTwoPartKey(ctx, conn, ownerDirID, sharedDirID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Shared Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Shared Directory (%s): %s", d.Id(), err)
	}

	d.Set("directory_id", output.OwnerDirectoryId)
	d.Set("method", output.ShareMethod)
	d.Set("notes", output.ShareNotes)
	d.Set("shared_directory_id", output.SharedDirectoryId)
	if output.SharedAccountId != nil {
		if err := d.Set(names.AttrTarget, []interface{}{flattenShareTarget(output)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target: %s", err)
		}
	} else {
		d.Set(names.AttrTarget, nil)
	}

	return diags
}

func resourceSharedDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	ownerDirID, sharedDirID, err := sharedDirectoryParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Directory Service Shared Directory: %s", d.Id())
	_, err = conn.UnshareDirectory(ctx, &directoryservice.UnshareDirectoryInput{
		DirectoryId:   aws.String(ownerDirID),
		UnshareTarget: expandUnshareTarget(d.Get(names.AttrTarget).([]interface{})[0].(map[string]interface{})),
	})

	if errs.IsA[*awstypes.DirectoryNotSharedException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Shared Directory (%s): %s", d.Id(), err)
	}

	if _, err := waitSharedDirectoryDeleted(ctx, conn, ownerDirID, sharedDirID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Directory Service Shared Directory (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const sharedDirectoryResourceIDSeparator = "/" // nosemgrep:ci.ds-in-const-name,ci.ds-in-var-name

func sharedDirectoryCreateResourceID(ownerDirectoryID, sharedDirectoryID string) string {
	parts := []string{ownerDirectoryID, sharedDirectoryID}
	id := strings.Join(parts, sharedDirectoryResourceIDSeparator)

	return id
}

func sharedDirectoryParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, sharedDirectoryResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected OWNER_DIRECTORY_ID%[2]sSHARED_DIRECTORY_ID", id, sharedDirectoryResourceIDSeparator)
}

func findSharedDirectory(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeSharedDirectoriesInput) (*awstypes.SharedDirectory, error) { // nosemgrep:ci.ds-in-func-name
	output, err := findSharedDirectories(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSharedDirectories(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeSharedDirectoriesInput) ([]awstypes.SharedDirectory, error) { // nosemgrep:ci.ds-in-func-name
	var output []awstypes.SharedDirectory

	pages := directoryservice.NewDescribeSharedDirectoriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SharedDirectories...)
	}

	return output, nil
}

func findSharedDirectoryByTwoPartKey(ctx context.Context, conn *directoryservice.Client, ownerDirectoryID, sharedDirectoryID string) (*awstypes.SharedDirectory, error) { // nosemgrep:ci.ds-in-func-name
	input := &directoryservice.DescribeSharedDirectoriesInput{
		OwnerDirectoryId:   aws.String(ownerDirectoryID),
		SharedDirectoryIds: []string{sharedDirectoryID},
	}

	output, err := findSharedDirectory(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.ShareStatus; status == awstypes.ShareStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusSharedDirectory(ctx context.Context, conn *directoryservice.Client, ownerDirectoryID, sharedDirectoryID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSharedDirectoryByTwoPartKey(ctx, conn, ownerDirectoryID, sharedDirectoryID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ShareStatus), nil
	}
}

func waitSharedDirectoryDeleted(ctx context.Context, conn *directoryservice.Client, ownerDirectoryID, sharedDirectoryID string, timeout time.Duration) (*awstypes.SharedDirectory, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ShareStatusDeleting,
			awstypes.ShareStatusShared,
			awstypes.ShareStatusPendingAcceptance,
			awstypes.ShareStatusRejectFailed,
			awstypes.ShareStatusRejected,
			awstypes.ShareStatusRejecting,
		),
		Target:                    []string{},
		Refresh:                   statusSharedDirectory(ctx, conn, ownerDirectoryID, sharedDirectoryID),
		Timeout:                   timeout,
		MinTimeout:                30 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SharedDirectory); ok {
		return output, err
	}

	return nil, err
}

func expandShareTarget(tfMap map[string]interface{}) *awstypes.ShareTarget { // nosemgrep:ci.ds-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ShareTarget{}

	if v, ok := tfMap[names.AttrID].(string); ok && len(v) > 0 {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && len(v) > 0 {
		apiObject.Type = awstypes.TargetType(v)
	}

	return apiObject
}

func expandUnshareTarget(tfMap map[string]interface{}) *awstypes.UnshareTarget {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.UnshareTarget{}

	if v, ok := tfMap[names.AttrID].(string); ok && len(v) > 0 {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && len(v) > 0 {
		apiObject.Type = awstypes.TargetType(v)
	}

	return apiObject
}

// flattenShareTarget is not a mirror of expandShareTarget because the API data structures are
// different, with no ShareTarget returned.
func flattenShareTarget(apiObject *awstypes.SharedDirectory) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SharedAccountId != nil {
		tfMap[names.AttrID] = aws.ToString(apiObject.SharedAccountId)
	}

	tfMap[names.AttrType] = awstypes.TargetTypeAccount // only type available

	return tfMap
}
