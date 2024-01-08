// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

//import (
//	"context"
//	"errors"
//	"fmt"
//	"log"
//	"time"
//
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
//	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	"github.com/hashicorp/terraform-provider-aws/internal/create"
//	"github.com/hashicorp/terraform-provider-aws/internal/errs"
//	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
//	"github.com/hashicorp/terraform-provider-aws/names"
//)
//
//// @SDKResource("aws_verifiedpermissions_policy_template", name="Policy Template")
//func ResourcePolicyTemplate() *schema.Resource {
//	return &schema.Resource{
//		CreateWithoutTimeout: resourcePolicyTemplateCreate,
//		ReadWithoutTimeout:   resourcePolicyTemplateRead,
//		UpdateWithoutTimeout: resourcePolicyTemplateUpdate,
//		DeleteWithoutTimeout: resourcePolicyTemplateDelete,
//
//		Importer: &schema.ResourceImporter{
//			StateContext: schema.ImportStatePassthroughContext,
//		},
//
//		Timeouts: &schema.ResourceTimeout{
//			Create: schema.DefaultTimeout(30 * time.Minute),
//			Update: schema.DefaultTimeout(30 * time.Minute),
//			Delete: schema.DefaultTimeout(30 * time.Minute),
//		},
//
//		Schema: map[string]*schema.Schema{
//			"policy_template_id": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"policy_store_id": {
//				Type:     schema.TypeString,
//				ForceNew: true,
//				Required: true,
//			},
//			"description": {
//				Type:     schema.TypeString,
//				Optional: true,
//			},
//			"statement": {
//				Type:     schema.TypeString,
//				Required: true,
//			},
//			"created_date": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"last_updated_date": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//		},
//	}
//}
//
//func resourcePolicyTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	in := &verifiedpermissions.CreatePolicyTemplateInput{
//		PolicyStoreId: aws.String(d.Get("policy_store_id").(string)),
//		Statement:     aws.String(d.Get("statement").(string)),
//		ClientToken:   aws.String(id.UniqueId()),
//	}
//
//	if description, ok := d.GetOk("description"); ok {
//		in.Description = aws.String(description.(string))
//	}
//
//	out, err := conn.CreatePolicyTemplate(ctx, in)
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyTemplate, d.Get("policy_store_id").(string), err)...)
//	}
//
//	if out == nil || out.PolicyStoreId == nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyTemplate, d.Get("policy_store_id").(string), errors.New("empty output"))...)
//	}
//
//	d.SetId(fmt.Sprintf("%s:%s", aws.ToString(out.PolicyStoreId), aws.ToString(out.PolicyTemplateId)))
//
//	return append(diags, resourcePolicyTemplateRead(ctx, d, meta)...)
//}
//
//func resourcePolicyTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	policyStoreID, policyTemplateID, err := policyTemplateParseID(d.Id())
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyTemplate, d.Id(), err)...)
//	}
//
//	out, err := findPolicyTemplateByID(ctx, conn, policyStoreID, policyTemplateID)
//
//	if !d.IsNewResource() && tfresource.NotFound(err) {
//		log.Printf("[WARN] VerifiedPermissions PolicyTemplate (%s) not found, removing from state", d.Id())
//		d.SetId("")
//		return diags
//	}
//
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyTemplate, d.Id(), err)...)
//	}
//
//	d.Set("policy_store_id", out.PolicyStoreId)
//	d.Set("created_date", out.CreatedDate.Format(time.RFC3339Nano))
//	d.Set("last_updated_date", out.LastUpdatedDate.Format(time.RFC3339Nano))
//	d.Set("description", out.Description)
//	d.Set("statement", out.Statement)
//	d.Set("policy_template_id", out.PolicyTemplateId)
//
//	return diags
//}
//
//func resourcePolicyTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	update := false
//
//	policyStoreId, policyTemplateID, err := policyTemplateParseID(d.Id())
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyTemplate, d.Id(), err)...)
//	}
//
//	in := &verifiedpermissions.UpdatePolicyTemplateInput{
//		PolicyStoreId:    aws.String(policyStoreId),
//		PolicyTemplateId: aws.String(policyTemplateID),
//	}
//
//	if d.HasChanges("statement") {
//		in.Statement = aws.String(d.Get("statement").(string))
//		update = true
//	}
//
//	if d.HasChanges("description") {
//		in.Description = aws.String(d.Get("description").(string))
//		update = true
//	}
//
//	if !update {
//		return diags
//	}
//
//	log.Printf("[DEBUG] Updating VerifiedPermissions PolicyTemplate (%s): %#v", d.Id(), in)
//	_, err = conn.UpdatePolicyTemplate(ctx, in)
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyTemplate, d.Id(), err)...)
//	}
//
//	return append(diags, resourcePolicyTemplateRead(ctx, d, meta)...)
//}
//
//func resourcePolicyTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).VerifiedPermissionsClient(ctx)
//
//	log.Printf("[INFO] Deleting VerifiedPermissions PolicyTemplate %s", d.Id())
//
//	policyStoreId, policyTemplateID, err := policyTemplateParseID(d.Id())
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicyTemplate, d.Id(), err)...)
//	}
//
//	_, err = conn.DeletePolicyTemplate(ctx, &verifiedpermissions.DeletePolicyTemplateInput{
//		PolicyStoreId:    aws.String(policyStoreId),
//		PolicyTemplateId: aws.String(policyTemplateID),
//	})
//
//	if errs.IsA[*types.ResourceNotFoundException](err) {
//		return diags
//	}
//	if err != nil {
//		return append(diags, create.DiagError(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicyTemplate, d.Id(), err)...)
//	}
//
//	return diags
//}
//
