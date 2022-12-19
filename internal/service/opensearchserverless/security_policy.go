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

func ResourceSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityPolicyCreate,
		ReadWithoutTimeout:   resourceSecurityPolicyRead,
		UpdateWithoutTimeout: resourceSecurityPolicyUpdate,
		DeleteWithoutTimeout: resourceSecurityPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				const idSeparator = "/"
				parts := strings.Split(d.Id(), idSeparator)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected security-policy-name%[2]ssecurity-policy-type", d.Id(), idSeparator)
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
				ValidateDiagFunc: enum.Validate[types.SecurityPolicyType](),
			},
		},
	}
}

const (
	ResNameSecurityPolicy = "Security Policy"
)

func resourceSecurityPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	in := &opensearchserverless.CreateSecurityPolicyInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(d.Get("name").(string)),
		Policy:      aws.String(d.Get("policy").(string)),
		Type:        types.SecurityPolicyType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateSecurityPolicy(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameSecurityPolicy, d.Get("name").(string), err)
	}

	if out == nil || out.SecurityPolicyDetail == nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionCreating, ResNameSecurityPolicy, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.SecurityPolicyDetail.Name))

	return resourceSecurityPolicyRead(ctx, d, meta)
}

func resourceSecurityPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	out, err := findSecurityPolicyByNameAndType(ctx, conn, d.Id(), d.Get("type").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearchServerless SecurityPolicy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionReading, ResNameSecurityPolicy, d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("type", out.Type)
	d.Set("description", out.Description)

	policyBytes, err := out.Policy.MarshalSmithyDocument()
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionSetting, ResNameSecurityPolicy, d.Id(), err)
	}

	p := string(policyBytes)

	p, err = verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), p)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionSetting, ResNameSecurityPolicy, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionSetting, ResNameSecurityPolicy, d.Id(), err)
	}

	d.Set("policy", p)

	return nil
}

func resourceSecurityPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	update := false

	in := &opensearchserverless.UpdateSecurityPolicyInput{
		ClientToken:   aws.String(resource.UniqueId()),
		Name:          aws.String(d.Id()),
		PolicyVersion: aws.String(d.Get("policy_version").(string)),
		Type:          types.SecurityPolicyType((d.Get("type").(string))),
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

	log.Printf("[DEBUG] Updating OpenSearchServerless SecurityPolicy (%s): %#v", d.Id(), in)
	_, err := conn.UpdateSecurityPolicy(ctx, in)
	if err != nil {
		return create.DiagError(names.OpenSearchServerless, create.ErrActionUpdating, ResNameSecurityPolicy, d.Id(), err)
	}

	return resourceSecurityPolicyRead(ctx, d, meta)
}

func resourceSecurityPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient()

	log.Printf("[INFO] Deleting OpenSearchServerless SecurityPolicy %s", d.Id())

	_, err := conn.DeleteSecurityPolicy(ctx, &opensearchserverless.DeleteSecurityPolicyInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(d.Id()),
		Type:        types.SecurityPolicyType(d.Get("type").(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.OpenSearchServerless, create.ErrActionDeleting, ResNameSecurityPolicy, d.Id(), err)
	}

	return nil
}
