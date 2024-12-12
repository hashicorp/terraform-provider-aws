// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_secretsmanager_secret", name="Secret")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/secretsmanager;secretsmanager.DescribeSecretOutput")
// @Testing(importIgnore="force_overwrite_replica_secret;recovery_window_in_days")
func resourceSecret() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretCreate,
		ReadWithoutTimeout:   resourceSecretRead,
		UpdateWithoutTimeout: resourceSecretUpdate,
		DeleteWithoutTimeout: resourceSecretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"force_overwrite_replica_secret": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validSecretName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validSecretNamePrefix,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"recovery_window_in_days": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
				ValidateFunc: validation.Any(
					validation.IntBetween(7, 30),
					validation.IntInSlice([]int{0}),
				),
			},
			"replica": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set: func(v interface{}) int {
					var buf bytes.Buffer

					m := v.(map[string]interface{})

					if v, ok := m[names.AttrKMSKeyID].(string); ok {
						buf.WriteString(fmt.Sprintf("%s-", v))
					}

					if v, ok := m[names.AttrRegion].(string); ok {
						buf.WriteString(fmt.Sprintf("%s-", v))
					}

					return create.StringHashcode(buf.String())
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"last_accessed_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrRegion: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &secretsmanager.CreateSecretInput{
		ClientRequestToken:          aws.String(id.UniqueId()), // Needed because we're handling our own retries
		Description:                 aws.String(d.Get(names.AttrDescription).(string)),
		ForceOverwriteReplicaSecret: d.Get("force_overwrite_replica_secret").(bool),
		Name:                        aws.String(name),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("replica"); ok && v.(*schema.Set).Len() > 0 {
		input.AddReplicaRegions = expandReplicaRegionTypes(v.(*schema.Set).List())
	}

	// Retry for secret recreation after deletion.
	outputRaw, err := tfresource.RetryWhen(ctx, PropagationTimeout,
		func() (interface{}, error) {
			return conn.CreateSecret(ctx, input)
		},
		func(err error) (bool, error) {
			// Temporarily retry on these errors to support immediate secret recreation:
			// InvalidRequestException: You can’t perform this operation on the secret because it was deleted.
			// InvalidRequestException: You can't create this secret because a secret with this name is already scheduled for deletion.
			if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "scheduled for deletion") || errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "was deleted") {
				return true, err
			}
			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Secrets Manager Secret (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*secretsmanager.CreateSecretOutput).ARN))

	_, err = tfresource.RetryWhenNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		return findSecretByID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Secrets Manager Secret (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok && v.(string) != "" && v.(string) != "{}" {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &secretsmanager.PutResourcePolicyInput{
			ResourcePolicy: aws.String(policy),
			SecretId:       aws.String(d.Id()),
		}

		if _, err := putSecretPolicy(ctx, conn, input); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceSecretRead(ctx, d, meta)...)
}

func resourceSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	output, err := findSecretByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Secrets Manager Secret (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ARN)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrKMSKeyID, output.KmsKeyId)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))
	if err := d.Set("replica", flattenReplicationStatusTypes(output.ReplicationStatus)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting replica: %s", err)
	}

	var policy *secretsmanager.GetResourcePolicyOutput
	err = tfresource.Retry(ctx, PropagationTimeout, func() *retry.RetryError {
		output, err := findSecretPolicyByID(ctx, conn, d.Id())

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if v := output.ResourcePolicy; v != nil {
			if valid, err := tfiam.PolicyHasValidAWSPrincipals(aws.ToString(v)); err != nil {
				return retry.NonRetryableError(err)
			} else if !valid {
				log.Printf("[DEBUG] Retrying because of invalid principals")
				return retry.RetryableError(errors.New("contains invalid principals"))
			}
		}

		policy = output

		return nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret (%s) policy: %s", d.Id(), err)
	} else if v := policy.ResourcePolicy; v != nil {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(v))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, "")
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	if d.HasChange("replica") {
		o, n := d.GetChange("replica")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if del := os.Difference(ns).List(); len(del) > 0 {
			if err := removeSecretReplicas(ctx, conn, d.Id(), expandReplicaRegionTypes(del)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if add := ns.Difference(os).List(); len(add) > 0 {
			if err := addSecretReplicas(ctx, conn, d.Id(), d.Get("force_overwrite_replica_secret").(bool), expandReplicaRegionTypes(add)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChanges(names.AttrDescription, names.AttrKMSKeyID) {
		input := &secretsmanager.UpdateSecretInput{
			ClientRequestToken: aws.String(id.UniqueId()), // Needed because we're handling our own retries
			Description:        aws.String(d.Get(names.AttrDescription).(string)),
			SecretId:           aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		_, err := conn.UpdateSecret(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPolicy) {
		if v, ok := d.GetOk(names.AttrPolicy); ok && v.(string) != "" && v.(string) != "{}" {
			policy, err := structure.NormalizeJsonString(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input := &secretsmanager.PutResourcePolicyInput{
				ResourcePolicy: aws.String(policy),
				SecretId:       aws.String(d.Id()),
			}

			if _, err := putSecretPolicy(ctx, conn, input); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			if err := deleteSecretPolicy(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceSecretRead(ctx, d, meta)...)
}

func resourceSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	if v, ok := d.GetOk("replica"); ok && v.(*schema.Set).Len() > 0 {
		if err := removeSecretReplicas(ctx, conn, d.Id(), expandReplicaRegionTypes(v.(*schema.Set).List())); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	input := &secretsmanager.DeleteSecretInput{
		SecretId: aws.String(d.Id()),
	}

	if v := d.Get("recovery_window_in_days").(int); v == 0 {
		input.ForceDeleteWithoutRecovery = aws.Bool(true)
	} else {
		input.RecoveryWindowInDays = aws.Int64(int64(v))
	}

	log.Printf("[DEBUG] Deleting Secrets Manager Secret: %s", d.Id())
	_, err := conn.DeleteSecret(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Secrets Manager Secret (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		return findSecretByID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Secrets Manager Secret (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func addSecretReplicas(ctx context.Context, conn *secretsmanager.Client, id string, forceOverwrite bool, replicas []types.ReplicaRegionType) error {
	if len(replicas) == 0 {
		return nil
	}

	input := &secretsmanager.ReplicateSecretToRegionsInput{
		AddReplicaRegions:           replicas,
		SecretId:                    aws.String(id),
		ForceOverwriteReplicaSecret: forceOverwrite,
	}

	_, err := conn.ReplicateSecretToRegions(ctx, input)

	if err != nil {
		return fmt.Errorf("adding Secrets Manager Secret (%s) replicas: %w", id, err)
	}

	return nil
}

func removeSecretReplicas(ctx context.Context, conn *secretsmanager.Client, id string, replicas []types.ReplicaRegionType) error {
	regions := tfslices.ApplyToAll(replicas, func(v types.ReplicaRegionType) string {
		return aws.ToString(v.Region)
	})

	if len(regions) == 0 {
		return nil
	}

	input := &secretsmanager.RemoveRegionsFromReplicationInput{
		RemoveReplicaRegions: regions,
		SecretId:             aws.String(id),
	}

	_, err := conn.RemoveRegionsFromReplication(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing Secrets Manager Secret (%s) replicas: %w", id, err)
	}

	return nil
}

func putSecretPolicy(ctx context.Context, conn *secretsmanager.Client, input *secretsmanager.PutResourcePolicyInput) (*secretsmanager.PutResourcePolicyOutput, error) {
	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.MalformedPolicyDocumentException](ctx, PropagationTimeout, func() (interface{}, error) {
		return conn.PutResourcePolicy(ctx, input)
	}, "This resource policy contains an unsupported principal")

	if err != nil {
		return nil, fmt.Errorf("putting Secrets Manager Secret (%s) policy: %w", aws.ToString(input.SecretId), err)
	}

	return outputRaw.(*secretsmanager.PutResourcePolicyOutput), nil
}

func deleteSecretPolicy(ctx context.Context, conn *secretsmanager.Client, id string) error {
	input := &secretsmanager.DeleteResourcePolicyInput{
		SecretId: aws.String(id),
	}

	_, err := conn.DeleteResourcePolicy(ctx, input)

	if err != nil {
		return fmt.Errorf("deleting Secrets Manager Secret (%s) policy: %w", id, err)
	}

	return nil
}

func findSecret(ctx context.Context, conn *secretsmanager.Client, input *secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error) {
	output, err := conn.DescribeSecret(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func findSecretByID(ctx context.Context, conn *secretsmanager.Client, id string) (*secretsmanager.DescribeSecretOutput, error) {
	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(id),
	}

	output, err := findSecret(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.DeletedDate != nil {
		return nil, &retry.NotFoundError{LastRequest: input}
	}

	return output, nil
}

func expandReplicaRegionType(tfMap map[string]interface{}) *types.ReplicaRegionType {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ReplicaRegionType{}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandReplicaRegionTypes(tfList []interface{}) []types.ReplicaRegionType {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ReplicaRegionType

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		if tfMap == nil {
			continue
		}

		apiObject := expandReplicaRegionType(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenReplicationStatusType(apiObject types.ReplicationStatusType) map[string]interface{} {
	tfMap := map[string]interface{}{
		names.AttrStatus: apiObject.Status,
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(v)
	}

	if v := apiObject.LastAccessedDate; v != nil {
		tfMap["last_accessed_date"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	if v := apiObject.StatusMessage; v != nil {
		tfMap[names.AttrStatusMessage] = aws.ToString(v)
	}

	return tfMap
}

func flattenReplicationStatusTypes(apiObjects []types.ReplicationStatusType) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenReplicationStatusType(apiObject))
	}

	return tfList
}
