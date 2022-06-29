package ds

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResourceNameSharedDirectory = "Shared Directory"
)

func ResourceSharedDirectory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSharedDirectoryCreate,
		ReadContext:   resourceSharedDirectoryRead,
		DeleteContext: resourceSharedDirectoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSharedDirectoryImport,
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
	conn := meta.(*conns.AWSClient).DSConn

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
		return names.DiagError(names.DS, names.ErrActionCreating, ResourceNameSharedDirectory, d.Id(), err)
	}

	log.Printf("[DEBUG] Shared Directory created: %s", out)
	d.SetId(fmt.Sprintf("%s/%s", dirId, aws.StringValue(out.SharedDirectoryId)))
	d.Set("shared_directory_id", out.SharedDirectoryId)

	return nil
}

func resourceSharedDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	dirId := d.Get("directory_id").(string)
	sharedId := d.Get("shared_directory_id").(string)

	output, err := findSharedDirectoryByIDs(ctx, conn, dirId, sharedId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.DS, names.ErrActionReading, ResourceNameSharedDirectory, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.DiagError(names.DS, names.ErrActionReading, ResourceNameSharedDirectory, d.Id(), err)
	}

	log.Printf("[DEBUG] Received DS shared directory: %s", output)

	d.Set("method", output.ShareMethod)
	d.Set("notes", output.ShareNotes)

	if output.SharedAccountId != nil {
		if err := d.Set("target", []interface{}{flattenShareTarget(output)}); err != nil {
			return names.DiagError(names.DS, names.ErrActionSetting, ResourceNameSharedDirectory, d.Id(), err)
		}
	} else {
		d.Set("target", nil)
	}

	return nil
}

func resourceSharedDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn

	dirId := d.Get("directory_id").(string)
	sharedId := d.Get("shared_directory_id").(string)

	input := directoryservice.UnshareDirectoryInput{
		DirectoryId:   aws.String(dirId),
		UnshareTarget: expandUnshareTarget(d.Get("target").([]interface{})[0].(map[string]interface{})),
	}

	// TODO: this takes forever and is not correctly waiting for unshare
	log.Printf("[DEBUG] Unsharing Directory Service Directory: %s", input)
	output, err := conn.UnshareDirectoryWithContext(ctx, &input)

	if err != nil {
		return names.DiagError(names.DS, names.ErrActionDeleting, ResourceNameSharedDirectory, d.Id(), err)
	}

	_, err = waitSharedDirectoryDeleted(ctx, conn, dirId, sharedId)

	if err != nil {
		return names.DiagError(names.DS, names.ErrActionWaitingForDeletion, ResourceNameSharedDirectory, d.Id(), err)
	}

	log.Printf("[DEBUG] Unshared Directory Service Directory: %s", output)

	return nil
}

func resourceSharedDirectoryImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <owner-directory-id>/<shared-directory-id>", d.Id())
	}

	ownerDirId := idParts[0]
	sharedDirId := idParts[1]

	d.Set("directory_id", ownerDirId)
	d.Set("shared_directory_id", sharedDirId)
	return []*schema.ResourceData{d}, nil
}

func expandShareTarget(tfMap map[string]interface{}) *directoryservice.ShareTarget { // nosemgrep:ds-in-func-name
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
