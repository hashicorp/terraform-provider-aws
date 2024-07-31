// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftserverless_resource_policy", name="Resource Policy")
func resourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	arn := d.Get(names.AttrResourceARN).(string)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := redshiftserverless.PutResourcePolicyInput{
		ResourceArn: aws.String(arn),
		Policy:      aws.String(policy),
	}

	out, err := conn.PutResourcePolicyWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Redshift Serverless Resource Policy (%s): %s", arn, err)
	}

	d.SetId(aws.StringValue(out.ResourcePolicy.ResourceArn))

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	out, err := findResourcePolicyByARN(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Resource Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceARN, out.ResourceArn)

	doc := resourcePolicyDoc{}
	log.Printf("policy is %s:", aws.StringValue(out.Policy))

	if err := json.Unmarshal([]byte(aws.StringValue(out.Policy)), &doc); err != nil {
		return sdkdiag.AppendErrorf(diags, "unmarshaling policy: %s", err)
	}

	doc.Statement.Resources = nil

	policyDoc := tfiam.IAMPolicyDoc{}

	policyDoc.Id = doc.Id
	policyDoc.Version = doc.Version
	policyDoc.Statements = []*tfiam.IAMPolicyStatement{doc.Statement}

	formattedPolicy, err := json.Marshal(policyDoc)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "marshling policy: %s", err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), string(formattedPolicy))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is an invalid JSON: %s", policyToSet, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Resource Policy: %s", d.Id())
	_, err := conn.DeleteResourcePolicyWithContext(ctx, &redshiftserverless.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

type resourcePolicyDoc struct {
	Version   string                    `json:",omitempty"`
	Id        string                    `json:",omitempty"`
	Statement *tfiam.IAMPolicyStatement `json:"Statement,omitempty"`
}
