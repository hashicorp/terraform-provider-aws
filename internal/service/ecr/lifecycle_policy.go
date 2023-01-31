package ecr

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLifecyclePolicyCreate,
		ReadWithoutTimeout:   resourceLifecyclePolicyRead,
		DeleteWithoutTimeout: resourceLifecyclePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentLifecyclePolicyJSON(old, new)

					return equal
				},
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLifecyclePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn()

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := &ecr.PutLifecyclePolicyInput{
		RepositoryName:      aws.String(d.Get("repository").(string)),
		LifecyclePolicyText: aws.String(policy),
	}

	resp, err := conn.PutLifecyclePolicyWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Lifecycle Policy (%s): %s", d.Get("repository").(string), err)
	}
	d.SetId(aws.StringValue(resp.RepositoryName))
	d.Set("registry_id", resp.RegistryId)
	return append(diags, resourceLifecyclePolicyRead(ctx, d, meta)...)
}

func resourceLifecyclePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn()

	input := &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(d.Id()),
	}

	var resp *ecr.GetLifecyclePolicyOutput

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.GetLifecyclePolicyWithContext(ctx, input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeLifecyclePolicyNotFoundException) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.GetLifecyclePolicyWithContext(ctx, input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeLifecyclePolicyNotFoundException) {
		log.Printf("[WARN] ECR Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
		log.Printf("[WARN] ECR Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Lifecycle Policy (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Lifecycle Policy (%s): empty response", d.Id())
	}

	d.Set("repository", resp.RepositoryName)
	d.Set("registry_id", resp.RegistryId)

	equivalent, err := equivalentLifecyclePolicyJSON(d.Get("policy").(string), aws.StringValue(resp.LifecyclePolicyText))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while comparing policy (state: %s) (from AWS: %s), encountered: %s", d.Get("policy").(string), aws.StringValue(resp.LifecyclePolicyText), err)
	}

	if !equivalent {
		policyToSet, err := structure.NormalizeJsonString(aws.StringValue(resp.LifecyclePolicyText))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyToSet, err)
		}

		d.Set("policy", policyToSet)
	}

	return diags
}

func resourceLifecyclePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn()

	input := &ecr.DeleteLifecyclePolicyInput{
		RepositoryName: aws.String(d.Id()),
	}

	_, err := conn.DeleteLifecyclePolicyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return diags
		}
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeLifecyclePolicyNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting ECR Lifecycle Policy (%s): %s", d.Id(), err)
	}

	return diags
}

type lifecyclePolicyRuleSelection struct {
	TagStatus     *string   `locationName:"tagStatus" type:"string" enum:"tagStatus" required:"true"`
	TagPrefixList []*string `locationName:"tagPrefixList" type:"list"`
	CountType     *string   `locationName:"countType" type:"string" enum:"countType" required:"true"`
	CountUnit     *string   `locationName:"countUnit" type:"string" enum:"countType"`
	CountNumber   *int64    `locationName:"countNumber" min:"1" type:"integer"`
}

type lifecyclePolicyRuleAction struct {
	ActionType *string `locationName:"type" type:"string" required:"true"`
}

type lifecyclePolicyRule struct {
	RulePriority *int64                        `locationName:"rulePriority" type:"integer" required:"true"`
	Description  *string                       `locationName:"description" type:"string"`
	Selection    *lifecyclePolicyRuleSelection `locationName:"selection" type:"structure" required:"true"`
	Action       *lifecyclePolicyRuleAction    `locationName:"action" type:"structure" required:"true"`
}

type lifecyclePolicy struct {
	Rules []*lifecyclePolicyRule `locationName:"rules" min:"1" type:"list" required:"true"`
}

func (lp *lifecyclePolicy) reduce() {
	sort.Slice(lp.Rules, func(i, j int) bool {
		return aws.Int64Value(lp.Rules[i].RulePriority) < aws.Int64Value(lp.Rules[j].RulePriority)
	})

	for _, rule := range lp.Rules {
		rule.Selection.reduce()
	}
}

func (lprs *lifecyclePolicyRuleSelection) reduce() {
	sort.Slice(lprs.TagPrefixList, func(i, j int) bool {
		return aws.StringValue(lprs.TagPrefixList[i]) < aws.StringValue(lprs.TagPrefixList[j])
	})

	if len(lprs.TagPrefixList) == 0 {
		lprs.TagPrefixList = nil
	}
}

func equivalentLifecyclePolicyJSON(str1, str2 string) (bool, error) {
	if strings.TrimSpace(str1) == "" {
		str1 = "{}"
	}

	if strings.TrimSpace(str2) == "" {
		str2 = "{}"
	}

	var lp1, lp2 lifecyclePolicy

	if err := json.Unmarshal([]byte(str1), &lp1); err != nil {
		return false, err
	}

	lp1.reduce()

	canonicalJSON1, err := jsonutil.BuildJSON(lp1)

	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(str2), &lp2); err != nil {
		return false, err
	}

	lp2.reduce()

	canonicalJSON2, err := jsonutil.BuildJSON(lp2)

	if err != nil {
		return false, err
	}

	equal := bytes.Equal(canonicalJSON1, canonicalJSON2)

	if !equal {
		log.Printf("[DEBUG] Canonical Lifecycle Policy JSONs are not equal.\nFirst: %s\nSecond: %s\n", canonicalJSON1, canonicalJSON2)
	}

	return equal, nil
}
