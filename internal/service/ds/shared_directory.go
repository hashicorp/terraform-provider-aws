// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameSharedDirectory = "Shared Directory"
)

// @SDKResource("aws_directory_service_shared_directory")
func ResourceSharedDirectory() *schema.Resource {
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      directoryservice.ShareMethodHandshake,
				ValidateFunc: validation.StringInSlice(directoryservice.ShareMethod_Values(), false),
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
			"target": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      directoryservice.TargetTypeAccount,
							ValidateFunc: validation.StringInSlice(directoryservice.TargetType_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceSharedDirectoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DSConn(ctx)

	dirId := d.Get("directory_id").(string)
	input := directoryservice.ShareDirectoryInput{
		DirectoryId: aws.String(dirId),
		ShareMethod: aws.String(d.Get("method").(string)),
		ShareTarget: expandShareTarget(d.Get("target").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk("notes"); ok {
		input.ShareNotes = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Shared Directory: %s", input)
	out, err := conn.ShareDirectoryWithContext(ctx, &input)

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionCreating, ResNameSharedDirectory, d.Id(), err)
	}

	log.Printf("[DEBUG] Shared Directory created: %s", out)
	d.SetId(sharedDirectoryID(dirId, aws.StringValue(out.SharedDirectoryId)))
	d.Set("shared_directory_id", out.SharedDirectoryId)

	return diags
}

func resourceSharedDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DSConn(ctx)

	ownerDirID, sharedDirID, err := parseSharedDirectoryID(d.Id())

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionReading, ResNameSharedDirectory, d.Id(), err)
	}

	output, err := FindSharedDirectory(ctx, conn, ownerDirID, sharedDirID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.DS, create.ErrActionReading, ResNameSharedDirectory, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionReading, ResNameSharedDirectory, d.Id(), err)
	}

	log.Printf("[DEBUG] Received DS shared directory: %s", output)

	d.Set("directory_id", output.OwnerDirectoryId)
	d.Set("method", output.ShareMethod)
	d.Set("notes", output.ShareNotes)
	d.Set("shared_directory_id", output.SharedDirectoryId)

	if output.SharedAccountId != nil {
		if err := d.Set("target", []interface{}{flattenShareTarget(output)}); err != nil {
			return create.AppendDiagError(diags, names.DS, create.ErrActionSetting, ResNameSharedDirectory, d.Id(), err)
		}
	} else {
		d.Set("target", nil)
	}

	return diags
}

func resourceSharedDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DSConn(ctx)

	dirId := d.Get("directory_id").(string)
	sharedId := d.Get("shared_directory_id").(string)

	input := directoryservice.UnshareDirectoryInput{
		DirectoryId:   aws.String(dirId),
		UnshareTarget: expandUnshareTarget(d.Get("target").([]interface{})[0].(map[string]interface{})),
	}

	log.Printf("[DEBUG] Unsharing Directory Service Directory: %s", input)
	output, err := conn.UnshareDirectoryWithContext(ctx, &input)

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionDeleting, ResNameSharedDirectory, d.Id(), err)
	}

	_, err = waitSharedDirectoryDeleted(ctx, conn, dirId, sharedId, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return create.AppendDiagError(diags, names.DS, create.ErrActionWaitingForDeletion, ResNameSharedDirectory, d.Id(), err)
	}

	log.Printf("[DEBUG] Unshared Directory Service Directory: %s", output)

	return diags
}

func expandShareTarget(tfMap map[string]interface{}) *directoryservice.ShareTarget { // nosemgrep:ci.ds-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.ShareTarget{}

	if v, ok := tfMap["id"].(string); ok && len(v) > 0 {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && len(v) > 0 {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandUnshareTarget(tfMap map[string]interface{}) *directoryservice.UnshareTarget {
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.UnshareTarget{}

	if v, ok := tfMap["id"].(string); ok && len(v) > 0 {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && len(v) > 0 {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

// flattenShareTarget is not a mirror of expandShareTarget because the API data structures are
// different, with no ShareTarget returned
func flattenShareTarget(apiObject *directoryservice.SharedDirectory) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SharedAccountId != nil {
		tfMap["id"] = aws.StringValue(apiObject.SharedAccountId)
	}

	tfMap["type"] = directoryservice.TargetTypeAccount // only type available

	return tfMap
}

func sharedDirectoryID(ownerDirectoryID, sharedDirectoryID string) string {
	return fmt.Sprintf("%s/%s", ownerDirectoryID, sharedDirectoryID)
}

func parseSharedDirectoryID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected <owner-directory-id>/<shared-directory-id>", id)
	}

	return idParts[0], idParts[1], nil
}
