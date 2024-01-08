// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_schemas_registry_policy")
func ResourceRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryPolicyCreate,
		ReadWithoutTimeout:   resourceRegistryPolicyRead,
		UpdateWithoutTimeout: resourceRegistryPolicyUpdate,
		DeleteWithoutTimeout: resourceRegistryPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	ResNameRegistryPolicy = "Registry Policy"
)

func resourceRegistryPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	registryName := d.Get("registry_name").(string)
	policy, err := structure.ExpandJsonFromString(d.Get("policy").(string))
	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionCreating, ResNameRegistryPolicy, registryName, err)
	}

	input := &schemas.PutResourcePolicyInput{
		Policy:       policy,
		RegistryName: aws.String(registryName),
	}

	log.Printf("[DEBUG] Creating EventBridge Schemas Registry Policy (%s)", d.Id())
	_, err = conn.PutResourcePolicyWithContext(ctx, input)

	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionCreating, ResNameRegistryPolicy, registryName, err)
	}

	d.SetId(registryName)

	return resourceRegistryPolicyRead(ctx, d, meta)
}

func resourceRegistryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	output, err := FindRegistryPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Registry Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionReading, ResNameRegistryPolicy, d.Id(), err)
	}

	policy, err := structure.FlattenJsonToString(output.Policy)
	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionReading, ResNameRegistryPolicy, d.Id(), err)
	}

	d.Set("policy", policy)
	d.Set("registry_name", d.Id())

	return nil
}

func resourceRegistryPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	policy, err := structure.ExpandJsonFromString(d.Get("policy").(string))
	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionUpdating, ResNameRegistryPolicy, d.Id(), err)
	}

	if d.HasChanges("policy") {
		input := &schemas.PutResourcePolicyInput{
			Policy:       policy,
			RegistryName: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Registry Policy (%s)", d.Id())
		_, err := conn.PutResourcePolicyWithContext(ctx, input)

		if err != nil {
			return create.DiagError(names.Schemas, create.ErrActionUpdating, ResNameRegistryPolicy, d.Id(), err)
		}
	}

	return resourceRegistryPolicyRead(ctx, d, meta)
}

func resourceRegistryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	log.Printf("[INFO] Deleting EventBridge Schemas Registry Policy (%s)", d.Id())
	_, err := conn.DeleteResourcePolicyWithContext(ctx, &schemas.DeleteResourcePolicyInput{
		RegistryName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionDeleting, ResNameRegistryPolicy, d.Id(), err)
	}

	return nil
}
