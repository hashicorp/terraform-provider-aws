// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_smb_file_share", name="SMB File Share")
// @Tags(identifierAttribute="arn")
func resourceSMBFileShare() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSMBFileShareCreate,
		ReadWithoutTimeout:   resourceSMBFileShareRead,
		UpdateWithoutTimeout: resourceSMBFileShareUpdate,
		DeleteWithoutTimeout: resourceSMBFileShareDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_based_enumeration": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"admin_user_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audit_destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"authentication": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      authenticationActiveDirectory,
				ValidateFunc: validation.StringInSlice(authentication_Values(), false),
			},
			"bucket_region": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"vpc_endpoint_dns_name"},
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(300, 2592000),
						},
					},
				},
			},
			"case_sensitivity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      storagegateway.CaseSensitivityClientSpecified,
				ValidateFunc: validation.StringInSlice(storagegateway.CaseSensitivity_Values(), false),
			},
			"default_storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultStorageClassS3Standard,
				ValidateFunc: validation.StringInSlice(defaultStorageClass_Values(), false),
			},
			"fileshare_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_share_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"guess_mime_type_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"invalid_user_list": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"kms_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
				RequiredWith: []string{"kms_encrypted"},
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"object_acl": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      storagegateway.ObjectACLPrivate,
				ValidateFunc: validation.StringInSlice(storagegateway.ObjectACL_Values(), false),
			},
			"oplocks_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"notification_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "{}",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^\{[\w\s:\{\}\[\]"]*}$`), ""),
					validation.StringLenBetween(2, 100),
				),
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"read_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"requester_pays": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"smb_acl_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"valid_user_list": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_endpoint_dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSMBFileShareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	input := &storagegateway.CreateSMBFileShareInput{
		AccessBasedEnumeration: aws.Bool(d.Get("access_based_enumeration").(bool)),
		ClientToken:            aws.String(id.UniqueId()),
		GatewayARN:             aws.String(d.Get("gateway_arn").(string)),
		GuessMIMETypeEnabled:   aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
		KMSEncrypted:           aws.Bool(d.Get("kms_encrypted").(bool)),
		LocationARN:            aws.String(d.Get("location_arn").(string)),
		ReadOnly:               aws.Bool(d.Get("read_only").(bool)),
		RequesterPays:          aws.Bool(d.Get("requester_pays").(bool)),
		Role:                   aws.String(d.Get(names.AttrRoleARN).(string)),
		SMBACLEnabled:          aws.Bool(d.Get("smb_acl_enabled").(bool)),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("admin_user_list"); ok && v.(*schema.Set).Len() > 0 {
		input.AdminUserList = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("authentication"); ok {
		input.Authentication = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bucket_region"); ok {
		input.BucketRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CacheAttributes = expandCacheAttributes(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("case_sensitivity"); ok {
		input.CaseSensitivity = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_storage_class"); ok {
		input.DefaultStorageClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_share_name"); ok {
		input.FileShareName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invalid_user_list"); ok && v.(*schema.Set).Len() > 0 {
		input.InvalidUserList = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_policy"); ok {
		input.NotificationPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_acl"); ok {
		input.ObjectACL = aws.String(v.(string))
	}

	if v, ok := d.GetOk("oplocks_enabled"); ok {
		input.OplocksEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("valid_user_list"); ok && v.(*schema.Set).Len() > 0 {
		input.ValidUserList = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("vpc_endpoint_dns_name"); ok {
		input.VPCEndpointDNSName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Storage Gateway SMB File Share: %s", input)
	output, err := conn.CreateSMBFileShareWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway SMB File Share: %s", err)
	}

	d.SetId(aws.StringValue(output.FileShareARN))

	if _, err = waitSMBFileShareCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway SMB File Share (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceSMBFileShareRead(ctx, d, meta)...)
}

func resourceSMBFileShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	fileshare, err := FindSMBFileShareByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway SMB File Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway SMB File Share (%s): %s", d.Id(), err)
	}

	d.Set("access_based_enumeration", fileshare.AccessBasedEnumeration)
	d.Set("admin_user_list", aws.StringValueSlice(fileshare.AdminUserList))
	d.Set(names.AttrARN, fileshare.FileShareARN)
	d.Set("audit_destination_arn", fileshare.AuditDestinationARN)
	d.Set("authentication", fileshare.Authentication)
	d.Set("bucket_region", fileshare.BucketRegion)

	if fileshare.CacheAttributes != nil {
		if err := d.Set("cache_attributes", []interface{}{flattenCacheAttributes(fileshare.CacheAttributes)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cache_attributes: %s", err)
		}
	} else {
		d.Set("cache_attributes", nil)
	}

	d.Set("case_sensitivity", fileshare.CaseSensitivity)
	d.Set("default_storage_class", fileshare.DefaultStorageClass)
	d.Set("fileshare_id", fileshare.FileShareId)
	d.Set("file_share_name", fileshare.FileShareName)
	d.Set("gateway_arn", fileshare.GatewayARN)
	d.Set("guess_mime_type_enabled", fileshare.GuessMIMETypeEnabled)
	d.Set("invalid_user_list", aws.StringValueSlice(fileshare.InvalidUserList))
	d.Set("kms_encrypted", fileshare.KMSEncrypted)
	d.Set(names.AttrKMSKeyARN, fileshare.KMSKey)
	d.Set("location_arn", fileshare.LocationARN)
	d.Set("notification_policy", fileshare.NotificationPolicy)
	d.Set("object_acl", fileshare.ObjectACL)
	d.Set("oplocks_enabled", fileshare.OplocksEnabled)
	d.Set(names.AttrPath, fileshare.Path)
	d.Set("read_only", fileshare.ReadOnly)
	d.Set("requester_pays", fileshare.RequesterPays)
	d.Set(names.AttrRoleARN, fileshare.Role)
	d.Set("smb_acl_enabled", fileshare.SMBACLEnabled)
	d.Set("valid_user_list", aws.StringValueSlice(fileshare.ValidUserList))
	d.Set("vpc_endpoint_dns_name", fileshare.VPCEndpointDNSName)

	setTagsOut(ctx, fileshare.Tags)

	return diags
}

func resourceSMBFileShareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &storagegateway.UpdateSMBFileShareInput{
			AccessBasedEnumeration: aws.Bool(d.Get("access_based_enumeration").(bool)),
			FileShareARN:           aws.String(d.Id()),
			GuessMIMETypeEnabled:   aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
			KMSEncrypted:           aws.Bool(d.Get("kms_encrypted").(bool)),
			ReadOnly:               aws.Bool(d.Get("read_only").(bool)),
			RequesterPays:          aws.Bool(d.Get("requester_pays").(bool)),
			SMBACLEnabled:          aws.Bool(d.Get("smb_acl_enabled").(bool)),
		}

		if d.HasChange("admin_user_list") {
			input.AdminUserList = flex.ExpandStringSet(d.Get("admin_user_list").(*schema.Set))
		}

		if d.HasChange("audit_destination_arn") {
			input.AuditDestinationARN = aws.String(d.Get("audit_destination_arn").(string))
		}

		if d.HasChange("cache_attributes") {
			if v, ok := d.GetOk("cache_attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.CacheAttributes = expandCacheAttributes(v.([]interface{})[0].(map[string]interface{}))
			} else {
				input.CacheAttributes = &storagegateway.CacheAttributes{}
			}
		}

		if d.HasChange("case_sensitivity") {
			input.CaseSensitivity = aws.String(d.Get("case_sensitivity").(string))
		}

		if d.HasChange("default_storage_class") {
			input.DefaultStorageClass = aws.String(d.Get("default_storage_class").(string))
		}

		if d.HasChange("file_share_name") {
			input.FileShareName = aws.String(d.Get("file_share_name").(string))
		}

		if d.HasChange("invalid_user_list") {
			input.InvalidUserList = flex.ExpandStringSet(d.Get("invalid_user_list").(*schema.Set))
		}

		// This value can only be set when KMSEncrypted is true.
		if d.HasChange(names.AttrKMSKeyARN) && d.Get("kms_encrypted").(bool) {
			input.KMSKey = aws.String(d.Get(names.AttrKMSKeyARN).(string))
		} else if d.Get("kms_encrypted").(bool) && d.Get(names.AttrKMSKeyARN).(string) != "" {
			input.KMSKey = aws.String(d.Get(names.AttrKMSKeyARN).(string))
		}

		if d.HasChange("notification_policy") {
			input.NotificationPolicy = aws.String(d.Get("notification_policy").(string))
		}

		if d.HasChange("object_acl") {
			input.ObjectACL = aws.String(d.Get("object_acl").(string))
		}

		if d.HasChange("oplocks_enabled") {
			input.OplocksEnabled = aws.Bool(d.Get("oplocks_enabled").(bool))
		}

		if d.HasChange("valid_user_list") {
			input.ValidUserList = flex.ExpandStringSet(d.Get("valid_user_list").(*schema.Set))
		}

		log.Printf("[DEBUG] Updating Storage Gateway SMB File Share: %s", input)
		_, err := conn.UpdateSMBFileShareWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway SMB File Share (%s): %s", d.Id(), err)
		}

		if _, err = waitSMBFileShareUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway SMB File Share (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSMBFileShareRead(ctx, d, meta)...)
}

func resourceSMBFileShareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway SMB File Share: %s", d.Id())
	_, err := conn.DeleteFileShareWithContext(ctx, &storagegateway.DeleteFileShareInput{
		FileShareARN: aws.String(d.Id()),
	})

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway SMB File Share (%s): %s", d.Id(), err)
	}

	if _, err = waitSMBFileShareDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway SMB File Share (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func expandCacheAttributes(tfMap map[string]interface{}) *storagegateway.CacheAttributes {
	if tfMap == nil {
		return nil
	}

	apiObject := &storagegateway.CacheAttributes{}

	if v, ok := tfMap["cache_stale_timeout_in_seconds"].(int); ok && v != 0 {
		apiObject.CacheStaleTimeoutInSeconds = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenCacheAttributes(apiObject *storagegateway.CacheAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CacheStaleTimeoutInSeconds; v != nil {
		tfMap["cache_stale_timeout_in_seconds"] = aws.Int64Value(v)
	}

	return tfMap
}
