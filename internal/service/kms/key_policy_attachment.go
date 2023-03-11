package kms

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_kms_key_policy_attachment")
func ResourceKeyPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceKeyPolicyAttachmentRead,
		UpdateWithoutTimeout: resourceKeyPolicyAttachmentUpdate,
		DeleteWithoutTimeout: resourceKeyPolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bypass_policy_lockout_safety_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceKeyPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	if err := updateKeyPolicy(ctx, conn, d.Get("key_id").(string), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching KMS Key policy (%s): %s", d.Id(), err)
	}

	d.SetId(d.Get("key_id").(string))

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	key, err := findKey(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", d.Id(), err)
	}

	d.Set("key_id", key.metadata.KeyId)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), key.policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", key.policy, err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceKeyPolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	if d.HasChange("policy") {
		if err := updateKeyPolicy(ctx, conn, d.Id(), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()
	accountId := meta.(*conns.AWSClient).AccountID

	if err := updateKeyPolicy(ctx, conn, d.Get("key_id").(string), defaultKeyPolicy(accountId), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching KMS Key policy (%s): %s", d.Id(), err)
	}

	return diags
}

func defaultKeyPolicy(accountId string) string {
	return fmt.Sprintf(`
{
	"Id": "default",
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "Enable IAM User Permissions",
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:aws:iam::%[1]s:root"
			},
			"Action": "kms:*",
			"Resource": "*"
		}
	]
}	
`, accountId)
}
