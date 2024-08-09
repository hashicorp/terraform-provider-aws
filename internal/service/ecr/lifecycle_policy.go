// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_lifecycle_policy", name="Lifecycle Policy")
func resourceLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLifecyclePolicyCreate,
		ReadWithoutTimeout:   resourceLifecyclePolicyRead,
		DeleteWithoutTimeout: resourceLifecyclePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentLifecyclePolicyJSON(old, new)
					return equal
				},
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLifecyclePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &ecr.PutLifecyclePolicyInput{
		LifecyclePolicyText: aws.String(policy),
		RepositoryName:      aws.String(d.Get("repository").(string)),
	}

	output, err := conn.PutLifecyclePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Lifecycle Policy (%s): %s", d.Get("repository").(string), err)
	}

	d.SetId(aws.ToString(output.RepositoryName))
	d.Set("registry_id", output.RegistryId)

	return append(diags, resourceLifecyclePolicyRead(ctx, d, meta)...)
}

func resourceLifecyclePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findLifecyclePolicyByRepositoryName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Lifecycle Policy (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*ecr.GetLifecyclePolicyOutput)

	if equivalent, err := equivalentLifecyclePolicyJSON(d.Get(names.AttrPolicy).(string), aws.ToString(output.LifecyclePolicyText)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else if !equivalent {
		policyToSet, err := structure.NormalizeJsonString(aws.ToString(output.LifecyclePolicyText))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	}

	d.Set("registry_id", output.RegistryId)
	d.Set("repository", output.RepositoryName)

	return diags
}

func resourceLifecyclePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Lifecycle Policy: %s", d.Id())
	_, err := conn.DeleteLifecyclePolicy(ctx, &ecr.DeleteLifecyclePolicyInput{
		RepositoryName: aws.String(d.Id()),
	})

	if errs.IsA[*types.LifecyclePolicyNotFoundException](err) || errs.IsA[*types.RepositoryNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Lifecycle Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findLifecyclePolicyByRepositoryName(ctx context.Context, conn *ecr.Client, repositoryName string) (*ecr.GetLifecyclePolicyOutput, error) {
	input := &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(repositoryName),
	}

	output, err := conn.GetLifecyclePolicy(ctx, input)

	if errs.IsA[*types.LifecyclePolicyNotFoundException](err) || errs.IsA[*types.RepositoryNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type lifecyclePolicyRuleSelection struct {
	TagStatus      *string   `json:"tagStatus,omitempty"`
	TagPatternList []*string `json:"tagPatternList,omitempty"`
	TagPrefixList  []*string `json:"tagPrefixList,omitempty"`
	CountType      *string   `json:"countType,omitempty"`
	CountUnit      *string   `json:"countUnit,omitempty"`
	CountNumber    *int64    `json:"countNumber,omitempty"`
}

type lifecyclePolicyRuleAction struct {
	Type *string `json:"type"`
}

type lifecyclePolicyRule struct {
	RulePriority *int64                        `json:"rulePriority,omitempty"`
	Description  *string                       `json:"description,omitempty"`
	Selection    *lifecyclePolicyRuleSelection `json:"selection,omitempty"`
	Action       *lifecyclePolicyRuleAction    `json:"action"`
}

type lifecyclePolicy struct {
	Rules []*lifecyclePolicyRule `json:"rules"`
}

func (lp *lifecyclePolicy) reduce() {
	sort.Slice(lp.Rules, func(i, j int) bool {
		return aws.ToInt64(lp.Rules[i].RulePriority) < aws.ToInt64(lp.Rules[j].RulePriority)
	})

	for _, rule := range lp.Rules {
		rule.Selection.reduce()
	}
}

func (lprs *lifecyclePolicyRuleSelection) reduce() {
	sort.Slice(lprs.TagPatternList, func(i, j int) bool {
		return aws.ToString(lprs.TagPatternList[i]) < aws.ToString(lprs.TagPatternList[j])
	})

	if len(lprs.TagPatternList) == 0 {
		lprs.TagPatternList = nil
	}

	sort.Slice(lprs.TagPrefixList, func(i, j int) bool {
		return aws.ToString(lprs.TagPrefixList[i]) < aws.ToString(lprs.TagPrefixList[j])
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

	var lp1 lifecyclePolicy
	err := tfjson.DecodeFromString(str1, &lp1)
	if err != nil {
		return false, err
	}
	lp1.reduce()
	b1, err := tfjson.EncodeToBytes(lp1)
	if err != nil {
		return false, err
	}

	var lp2 lifecyclePolicy
	err = tfjson.DecodeFromString(str2, &lp2)
	if err != nil {
		return false, err
	}
	lp2.reduce()
	b2, err := tfjson.EncodeToBytes(lp2)
	if err != nil {
		return false, err
	}

	return tfjson.EqualBytes(b1, b2), nil
}
