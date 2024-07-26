// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_access_point")
func resourceAccessPoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPointCreate,
		ReadWithoutTimeout:   resourceAccessPointRead,
		UpdateWithoutTimeout: resourceAccessPointUpdate,
		DeleteWithoutTimeout: resourceAccessPointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"bucket_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoints: {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"has_public_access_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"network_origin": {
				Type:     schema.TypeString,
				Computed: true,
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
			"public_access_block_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MinItems:         0,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"block_public_acls": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"block_public_policy": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"ignore_public_acls": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"restrict_public_buckets": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrVPCConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceAccessPointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	name := d.Get(names.AttrName).(string)
	input := &s3control.CreateAccessPointInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(d.Get(names.AttrBucket).(string)),
		Name:      aws.String(name),
	}

	if v, ok := d.GetOk("bucket_account_id"); ok {
		input.BucketAccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_access_block_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.PublicAccessBlockConfiguration = expandPublicAccessBlockConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrVPCConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VpcConfiguration = expandVPCConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateAccessPoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Access Point (%s): %s", name, err)
	}

	resourceID, err := AccessPointCreateResourceID(aws.ToString(output.AccessPointArn))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	accountID, name, err = AccessPointParseResourceID(resourceID)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(resourceID)

	if v, ok := d.GetOk(names.AttrPolicy); ok && v.(string) != "" && v.(string) != "{}" {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &s3control.PutAccessPointPolicyInput{
			AccountId: aws.String(accountID),
			Name:      aws.String(name),
			Policy:    aws.String(policy),
		}

		_, err = conn.PutAccessPointPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating S3 Access Point (%s) policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAccessPointRead(ctx, d, meta)...)
}

func resourceAccessPointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := AccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findAccessPointByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Access Point (%s): %s", d.Id(), err)
	}

	s3OnOutposts := arn.IsARN(name)

	if s3OnOutposts {
		accessPointARN, err := arn.Parse(name)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazons3onoutposts.html#amazons3onoutposts-resources-for-iam-policies.
		bucketARN := arn.ARN{
			Partition: accessPointARN.Partition,
			Service:   accessPointARN.Service,
			Region:    accessPointARN.Region,
			AccountID: accessPointARN.AccountID,
			Resource: strings.Replace(
				accessPointARN.Resource,
				fmt.Sprintf("accesspoint/%s", aws.ToString(output.Name)),
				fmt.Sprintf("bucket/%s", aws.ToString(output.Bucket)),
				1,
			),
		}

		d.Set(names.AttrARN, name)
		d.Set(names.AttrBucket, bucketARN.String())
	} else {
		// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazons3.html#amazons3-resources-for-iam-policies.
		accessPointARN := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "s3",
			Region:    meta.(*conns.AWSClient).Region,
			AccountID: accountID,
			Resource:  fmt.Sprintf("accesspoint/%s", aws.ToString(output.Name)),
		}

		d.Set(names.AttrARN, accessPointARN.String())
		d.Set(names.AttrBucket, output.Bucket)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrAlias, output.Alias)
	d.Set("bucket_account_id", output.BucketAccountId)
	d.Set(names.AttrDomainName, meta.(*conns.AWSClient).RegionalHostname(ctx, fmt.Sprintf("%s-%s.s3-accesspoint", aws.ToString(output.Name), accountID)))
	d.Set(names.AttrEndpoints, output.Endpoints)
	d.Set(names.AttrName, output.Name)
	d.Set("network_origin", output.NetworkOrigin)
	if output.PublicAccessBlockConfiguration != nil {
		if err := d.Set("public_access_block_configuration", []interface{}{flattenPublicAccessBlockConfiguration(output.PublicAccessBlockConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting public_access_block_configuration: %s", err)
		}
	} else {
		d.Set("public_access_block_configuration", nil)
	}
	if output.VpcConfiguration != nil {
		if err := d.Set(names.AttrVPCConfiguration, []interface{}{flattenVPCConfiguration(output.VpcConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_configuration: %s", err)
		}
	} else {
		d.Set(names.AttrVPCConfiguration, nil)
	}

	policy, status, err := findAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

	if err == nil && policy != "" {
		if s3OnOutposts {
			d.Set("has_public_access_policy", false)
		} else {
			d.Set("has_public_access_policy", status.IsPublic)
		}

		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else if policy == "" || tfresource.NotFound(err) {
		d.Set("has_public_access_policy", false)
		d.Set(names.AttrPolicy, nil)
	} else {
		return sdkdiag.AppendErrorf(diags, "reading S3 Access Point (%s) policy: %s", d.Id(), err)
	}

	return diags
}

func resourceAccessPointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := AccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChange(names.AttrPolicy) {
		if v, ok := d.GetOk(names.AttrPolicy); ok && v.(string) != "" && v.(string) != "{}" {
			policy, err := structure.NormalizeJsonString(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input := &s3control.PutAccessPointPolicyInput{
				AccountId: aws.String(accountID),
				Name:      aws.String(name),
				Policy:    aws.String(policy),
			}

			_, err = conn.PutAccessPointPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating S3 Access Point (%s) policy: %s", d.Id(), err)
			}
		} else {
			input := &s3control.DeleteAccessPointPolicyInput{
				AccountId: aws.String(accountID),
				Name:      aws.String(name),
			}

			_, err := conn.DeleteAccessPointPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Access Point (%s) policy: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceAccessPointRead(ctx, d, meta)...)
}

func resourceAccessPointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := AccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Access Point: %s", d.Id())
	_, err = conn.DeleteAccessPoint(ctx, &s3control.DeleteAccessPointInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Access Point (%s): %s", d.Id(), err)
	}

	return diags
}

func findAccessPointByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (*s3control.GetAccessPointOutput, error) {
	input := &s3control.GetAccessPointInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetAccessPoint(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
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

const accessPointResourceIDSeparator = ":"

func AccessPointCreateResourceID(accessPointARN string) (string, error) {
	v, err := arn.Parse(accessPointARN)

	if err != nil {
		return "", err
	}

	switch service := v.Service; service {
	case "s3":
		resource := v.Resource
		if !strings.HasPrefix(resource, "accesspoint/") {
			return "", fmt.Errorf("unexpected resource: %s", resource)
		}

		parts := []string{v.AccountID, strings.TrimPrefix(resource, "accesspoint/")}
		id := strings.Join(parts, accessPointResourceIDSeparator)

		return id, nil

	case "s3-outposts":
		return accessPointARN, nil

	default:
		return "", fmt.Errorf("unexpected service: %s", service)
	}
}

func AccessPointParseResourceID(id string) (string, string, error) {
	if v, err := arn.Parse(id); err == nil {
		return v.AccountID, id, nil
	}

	parts := strings.Split(id, accessPointResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected account-id%[2]saccess-point-name", id, accessPointResourceIDSeparator)
}

func expandVPCConfiguration(tfMap map[string]interface{}) *types.VpcConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VpcConfiguration{}

	if v, ok := tfMap[names.AttrVPCID].(string); ok {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenVPCConfiguration(apiObject *types.VpcConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.VpcId; v != nil {
		tfMap[names.AttrVPCID] = aws.ToString(v)
	}

	return tfMap
}
