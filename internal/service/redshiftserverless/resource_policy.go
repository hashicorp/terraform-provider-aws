// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshiftserverless

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrPolicy: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateFunc:     validation.StringIsJSON,
					DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
					StateFunc: func(v any) string {
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
			}
		},
	}
}

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	arn := d.Get(names.AttrResourceARN).(string)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := &redshiftserverless.PutResourcePolicyInput{
		ResourceArn: aws.String(arn),
		Policy:      aws.String(policy),
	}

	out, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Redshift Serverless Resource Policy (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(out.ResourcePolicy.ResourceArn))

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	out, err := findResourcePolicyByARN(ctx, conn, d.Id())
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Resource Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceARN, out.ResourceArn)

	doc := resourcePolicyDoc{}
	log.Printf("policy is %s:", aws.ToString(out.Policy))

	if err := json.Unmarshal([]byte(aws.ToString(out.Policy)), &doc); err != nil {
		return sdkdiag.AppendErrorf(diags, "unmarshaling policy: %s", err)
	}

	for _, statement := range doc.Statement {
		statement.Resources = nil
	}

	policyDoc := tfiam.IAMPolicyDoc{}

	policyDoc.Id = doc.Id
	policyDoc.Version = doc.Version
	policyDoc.Statements = doc.Statement

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

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Resource Policy: %s", d.Id())
	_, err := conn.DeleteResourcePolicy(ctx, &redshiftserverless.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicyByARN(ctx context.Context, conn *redshiftserverless.Client, arn string) (*awstypes.ResourcePolicy, error) {
	input := &redshiftserverless.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.ResourceNotFoundException](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ResourcePolicy, nil
}

type resourcePolicyDoc struct {
	Version   string                      `json:",omitempty"`
	Id        string                      `json:",omitempty"`
	Statement []*tfiam.IAMPolicyStatement `json:"Statement,omitempty"`
}

// UnmarshalJSON allows the Statement field to be either a single statement
// object or an array of statement objects, matching the IAM policy JSON
// grammar. The AWS API returns Statement as an array, so unmarshaling into a
// single struct field previously failed with "cannot unmarshal array into Go
// struct field resourcePolicyDoc.Statement".
func (d *resourcePolicyDoc) UnmarshalJSON(b []byte) error {
	// Alias to avoid recursing into this method.
	type resourcePolicyDocAlias struct {
		Version   string          `json:",omitempty"`
		Id        string          `json:",omitempty"`
		Statement json.RawMessage `json:"Statement,omitempty"`
	}

	var alias resourcePolicyDocAlias
	if err := json.Unmarshal(b, &alias); err != nil {
		return err
	}

	d.Version = alias.Version
	d.Id = alias.Id

	if len(alias.Statement) == 0 {
		return nil
	}

	// Try to unmarshal as an array first, then fall back to a single object.
	if err := json.Unmarshal(alias.Statement, &d.Statement); err != nil {
		var single *tfiam.IAMPolicyStatement
		if err := json.Unmarshal(alias.Statement, &single); err != nil {
			return err
		}
		d.Statement = []*tfiam.IAMPolicyStatement{single}
	}

	return nil
}
