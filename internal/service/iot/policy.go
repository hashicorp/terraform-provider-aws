// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_policy")
// @Tags(identifierAttribute="arn")
func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrPolicy: {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	name := d.Get(names.AttrName).(string)
	input := &iot.CreatePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreatePolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.PolicyName))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	output, err := FindPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.PolicyArn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set(names.AttrName, output.PolicyName)

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.StringValue(output.PolicyDocument))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
		}

		input := &iot.CreatePolicyVersionInput{
			PolicyDocument: aws.String(policy),
			PolicyName:     aws.String(d.Id()),
			SetAsDefault:   aws.Bool(true),
		}

		_, errCreate := conn.CreatePolicyVersionWithContext(ctx, input)

		// "VersionsLimitExceededException: The policy ... already has the maximum number of versions (5)"
		if tfawserr.ErrCodeEquals(errCreate, iot.ErrCodeVersionsLimitExceededException) {
			// Prune the lowest version and retry.
			policyVersions, err := FindPolicyVersionsByName(ctx, conn, d.Id())

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s) versions: %s", d.Id(), err)
			}

			var versionIDs []int

			for _, v := range policyVersions {
				if aws.BoolValue(v.IsDefaultVersion) {
					continue
				}

				v, err := strconv.Atoi(aws.StringValue(v.VersionId))

				if err != nil {
					continue
				}

				versionIDs = append(versionIDs, v)
			}

			if len(versionIDs) > 0 {
				// Sort ascending.
				slices.Sort(versionIDs)
				versionID := strconv.Itoa(versionIDs[0])

				if err := deletePolicyVersion(ctx, conn, d.Id(), versionID, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for IoT Policy (%s) version (%s) delete: %s", d.Id(), versionID, err)
				}

				_, errCreate = conn.CreatePolicyVersionWithContext(ctx, input)
			}
		}

		if errCreate != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Policy (%s): %s", d.Id(), errCreate)
		}
	}
	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	policyVersions, err := FindPolicyVersionsByName(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s) versions: %s", d.Id(), err)
	}

	// Delete all non-default versions of the policy.
	for _, v := range policyVersions {
		if aws.BoolValue(v.IsDefaultVersion) {
			continue
		}

		if err := deletePolicyVersion(ctx, conn, d.Id(), aws.StringValue(v.VersionId), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	// Delete default policy version.
	if err := deletePolicy(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func FindPolicyByName(ctx context.Context, conn *iot.IoT, name string) (*iot.GetPolicyOutput, error) {
	input := &iot.GetPolicyInput{
		PolicyName: aws.String(name),
	}

	output, err := conn.GetPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
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

func FindPolicyVersionsByName(ctx context.Context, conn *iot.IoT, name string) ([]*iot.PolicyVersion, error) {
	input := &iot.ListPolicyVersionsInput{
		PolicyName: aws.String(name),
	}

	output, err := conn.ListPolicyVersionsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PolicyVersions) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PolicyVersions, nil
}

func deletePolicy(ctx context.Context, conn *iot.IoT, name string, timeout time.Duration) error {
	input := &iot.DeletePolicyInput{
		PolicyName: aws.String(name),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.DeletePolicyWithContext(ctx, input)
	}, iot.ErrCodeDeleteConflictException)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IoT Policy (%s): %w", name, err)
	}

	return nil
}

func deletePolicyVersion(ctx context.Context, conn *iot.IoT, name, versionID string, timeout time.Duration) error {
	input := &iot.DeletePolicyVersionInput{
		PolicyName:      aws.String(name),
		PolicyVersionId: aws.String(versionID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.DeletePolicyVersionWithContext(ctx, input)
	}, iot.ErrCodeDeleteConflictException)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IoT Policy (%s) version (%s): %w", name, versionID, err)
	}

	return nil
}
