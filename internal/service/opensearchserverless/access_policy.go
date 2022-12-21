package opensearchserverless

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceAccessPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPolicyCreate,
		ReadWithoutTimeout:   resourceAccessPolicyRead,
		UpdateWithoutTimeout: resourceAccessPolicyUpdate,
		DeleteWithoutTimeout: resourceAccessPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				const idSeparator = "/"
				parts := strings.Split(d.Id(), idSeparator)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected access-policy-name%[2]saccess-policy-type", d.Id(), idSeparator)
				}

				d.SetId(parts[0])
				d.Set("type", parts[1])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(3, 32),
			},
			"policy": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					validation.StringLenBetween(1, 20480),
				),
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"policy_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.AccessPolicyType](),
			},
		},
	}
}

const (
	ResNameAccessPolicy = "Access Policy"
)

func resourceAccessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	in := &opensearchserverless.CreateAccessPolicyInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(d.Get("name").(string)),
		Policy:      aws.String(d.Get("policy").(string)),
		Type:        types.AccessPolicyType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateAccessPolicy(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameAccessPolicy, d.Get("name").(string), err)
	}

	if out == nil || out.AccessPolicyDetail == nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameAccessPolicy, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.AccessPolicyDetail.Name))

	return resourceAccessPolicyRead(ctx, d, meta)
}

func resourceAccessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()
	out, err := findAccessPolicyByNameAndType(ctx, conn, d.Id(), d.Get("type").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearchServerless AccessPolicy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionReading, ResNameAccessPolicy, d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("type", out.Type)
	d.Set("description", out.Description)

	policyBytes, err := out.Policy.MarshalSmithyDocument()
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionSetting, ResNameAccessPolicy, d.Id(), err)
	}

	p := string(policyBytes)

	p, err = verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), p)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionSetting, ResNameAccessPolicy, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionSetting, ResNameAccessPolicy, d.Id(), err)
	}

	d.Set("policy", p)

	return nil
}

func resourceAccessPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	update := false

	in := &opensearchserverless.UpdateAccessPolicyInput{
		ClientToken:   aws.String(resource.UniqueId()),
		Name:          aws.String(d.Id()),
		PolicyVersion: aws.String(d.Get("policy_version").(string)),
		Type:          types.AccessPolicyType((d.Get("type").(string))),
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChanges("policy") {
		in.Policy = aws.String(d.Get("policy").(string))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating OpenSearchServerless AccessPolicy (%s): %#v", d.Id(), in)
	_, err := conn.UpdateAccessPolicy(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionUpdating, ResNameAccessPolicy, d.Id(), err)
	}

	return resourceAccessPolicyRead(ctx, d, meta)
}

func resourceAccessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	log.Printf("[INFO] Deleting OpenSearchServerless AccessPolicy %s", d.Id())

	_, err := conn.DeleteAccessPolicy(ctx, &opensearchserverless.DeleteAccessPolicyInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(d.Id()),
		Type:        types.AccessPolicyType(d.Get("type").(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.OpenSearchServerless, create.ErrActionDeleting, ResNameAccessPolicy, d.Id(), err)
	}

	return nil
}
