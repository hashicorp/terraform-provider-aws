package schemas

import (
	"context"
	"log"
	"time"

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

func ResourceRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryPolicyCreate,
		ReadWithoutTimeout:   resourceRegistryPolicyRead,
		UpdateWithoutTimeout: resourceRegistryPolicyUpdate,
		DeleteWithoutTimeout: resourceRegistryPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

const (
	ResNameRegistryPolicy = "Registry Policy"
)

func resourceRegistryPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn

	registryName := d.Get("registry_name").(string)
	policy, err := structure.ExpandJsonFromString(d.Get("policy").(string))

	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionCreating, ResNameRegistryPolicy, registryName, err)
	}

	input := schemas.PutResourcePolicyInput{
		RegistryName: aws.String(registryName),
		Policy:       policy,
	}

	log.Printf("[DEBUG] Creating EventBridge Schemas Registry Policy (%s)", d.Id())
	_, err = conn.PutResourcePolicy(&input)

	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionCreating, ResNameRegistryPolicy, registryName, err)
	}

	d.SetId(aws.StringValue(input.RegistryName))
	return resourceRegistryPolicyRead(ctx, d, meta)
}

func resourceRegistryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn

	input := &schemas.GetResourcePolicyInput{
		RegistryName: aws.String(d.Id()),
	}

	output, err := conn.GetResourcePolicy(input)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Registry Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	policy, _ := structure.FlattenJsonToString(output.Policy)
	d.Set("registry_name", input.RegistryName)
	d.Set("policy", policy)
	return nil
}

func resourceRegistryPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn
	policy, err := structure.ExpandJsonFromString(d.Get("policy").(string))

	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionUpdating, ResNameRegistryPolicy, d.Id(), err)
	}

	if d.HasChanges("policy") {
		input := &schemas.PutResourcePolicyInput{
			RegistryName: aws.String(d.Id()),
			Policy:       policy,
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Registry Policy (%s)", d.Id())
		_, err := conn.PutResourcePolicy(input)

		if err != nil {
			return create.DiagError(names.Schemas, create.ErrActionUpdating, ResNameRegistryPolicy, d.Id(), err)
		}
	}

	return resourceRegistryPolicyRead(ctx, d, meta)
}

func resourceRegistryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SchemasConn
	input := &schemas.DeleteResourcePolicyInput{
		RegistryName: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EventBridge Schemas Registry Policy (%s)", d.Id())
	_, err := conn.DeleteResourcePolicy(input)

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Schemas, create.ErrActionDeleting, ResNameRegistryPolicy, d.Id(), err)
	}

	return nil
}
