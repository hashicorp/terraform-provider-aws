// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_compute_environment", name="Compute Environment")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/batch/types;types.ComputeEnvironmentDetail")
func resourceComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceComputeEnvironmentCreate,
		ReadWithoutTimeout:   resourceComputeEnvironmentRead,
		UpdateWithoutTimeout: resourceComputeEnvironmentUpdate,
		DeleteWithoutTimeout: resourceComputeEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: resourceComputeEnvironmentCustomizeDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_environment_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name_prefix"},
				ValidateFunc:  validName,
			},
			"compute_environment_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name"},
				ValidateFunc:  validPrefix,
			},
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:             schema.TypeString,
							Optional:         true,
							StateFunc:        sdkv2.ToUpperSchemaStateFunc,
							ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.CRAllocationStrategy](),
						},
						"bid_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"desired_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"ec2_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_id_override": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"image_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"ec2_key_pair": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"image_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"instance_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrInstanceType: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrLaunchTemplate: {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
									},
									"launch_template_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
									},
									names.AttrVersion: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"max_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"min_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"placement_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"spot_iam_fleet_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrTags: tftags.TagsSchema(),
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							StateFunc:        sdkv2.ToUpperSchemaStateFunc,
							ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.CRType](),
						},
					},
				},
			},
			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"eks_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"eks_cluster_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"kubernetes_namespace": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrServiceRole: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrState: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.CEStateEnabled,
				StateFunc:        sdkv2.ToUpperSchemaStateFunc,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.CEState](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				StateFunc:        sdkv2.ToUpperSchemaStateFunc,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.CEType](),
			},
			"update_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_execution_timeout_minutes": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 360),
						},
						"terminate_jobs_on_update": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceComputeEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	computeEnvironmentName := create.Name(d.Get("compute_environment_name").(string), d.Get("compute_environment_name_prefix").(string))
	computeEnvironmentType := awstypes.CEType(d.Get(names.AttrType).(string))
	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(computeEnvironmentName),
		ServiceRole:            aws.String(d.Get(names.AttrServiceRole).(string)),
		Tags:                   getTagsIn(ctx),
		Type:                   computeEnvironmentType,
	}

	if v, ok := d.GetOk("compute_resources"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ComputeResources = expandComputeResource(ctx, v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("eks_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.EksConfiguration = expandEKSConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrState); ok {
		input.State = awstypes.CEState(v.(string))
	}

	output, err := conn.CreateComputeEnvironment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Batch Compute Environment (%s): %s", computeEnvironmentName, err)
	}

	d.SetId(aws.ToString(output.ComputeEnvironmentName))

	if _, err := waitComputeEnvironmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) create: %s", d.Id(), err)
	}

	// UpdatePolicy is not possible to set with CreateComputeEnvironment
	if v, ok := d.GetOk("update_policy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
			UpdatePolicy:       expandComputeEnvironmentUpdatePolicy(v.([]any)),
		}

		_, err := conn.UpdateComputeEnvironment(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Compute Environment (%s) update policy: %s", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceComputeEnvironmentRead(ctx, d, meta)...)
}

func resourceComputeEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	computeEnvironment, err := findComputeEnvironmentDetailByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Compute Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("compute_environment_name_prefix", create.NamePrefixFromName(aws.ToString(computeEnvironment.ComputeEnvironmentName)))
	if computeEnvironment.ComputeResources != nil {
		if err := d.Set("compute_resources", []any{flattenComputeResource(ctx, computeEnvironment.ComputeResources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting compute_resources: %s", err)
		}
	} else {
		d.Set("compute_resources", nil)
	}
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	if computeEnvironment.EksConfiguration != nil {
		if err := d.Set("eks_configuration", []any{flattenEKSConfiguration(computeEnvironment.EksConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting eks_configuration: %s", err)
		}
	} else {
		d.Set("eks_configuration", nil)
	}
	d.Set(names.AttrServiceRole, computeEnvironment.ServiceRole)
	d.Set(names.AttrState, computeEnvironment.State)
	d.Set(names.AttrStatus, computeEnvironment.Status)
	d.Set(names.AttrStatusReason, computeEnvironment.StatusReason)
	d.Set(names.AttrType, computeEnvironment.Type)
	if err := d.Set("update_policy", flattenComputeEnvironmentUpdatePolicy(computeEnvironment.UpdatePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting update_policy: %s", err)
	}

	setTagsOut(ctx, computeEnvironment.Tags)

	return diags
}

func resourceComputeEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrServiceRole) {
			input.ServiceRole = aws.String(d.Get(names.AttrServiceRole).(string))
		}

		if d.HasChange(names.AttrState) {
			input.State = awstypes.CEState(d.Get(names.AttrState).(string))
		}

		if d.HasChange("update_policy") {
			input.UpdatePolicy = expandComputeEnvironmentUpdatePolicy(d.Get("update_policy").([]any))
		}

		if computeEnvironmentType := strings.ToUpper(d.Get(names.AttrType).(string)); computeEnvironmentType == string(awstypes.CETypeManaged) {
			// "At least one compute-resources attribute must be specified"
			computeResourceUpdate := &awstypes.ComputeResourceUpdate{
				MaxvCpus: aws.Int32(int32(d.Get("compute_resources.0.max_vcpus").(int))),
			}

			if d.HasChange("compute_resources.0.security_group_ids") {
				computeResourceUpdate.SecurityGroupIds = flex.ExpandStringValueSet(d.Get("compute_resources.0.security_group_ids").(*schema.Set))
			}

			if d.HasChange("compute_resources.0.subnets") {
				computeResourceUpdate.Subnets = flex.ExpandStringValueSet(d.Get("compute_resources.0.subnets").(*schema.Set))
			}

			if d.HasChange("compute_resources.0.allocation_strategy") {
				if allocationStrategy, ok := d.GetOk("compute_resources.0.allocation_strategy"); ok {
					computeResourceUpdate.AllocationStrategy = awstypes.CRUpdateAllocationStrategy(allocationStrategy.(string))
				} else {
					computeResourceUpdate.AllocationStrategy = ""
				}
			}

			computeResourceEnvironmentType := awstypes.CRType(d.Get("compute_resources.0.type").(string))

			if d.HasChange("compute_resources.0.type") {
				computeResourceUpdate.Type = computeResourceEnvironmentType
			}

			if !isFargateType(computeResourceEnvironmentType) {
				if d.HasChange("compute_resources.0.desired_vcpus") {
					if desiredvCpus, ok := d.GetOk("compute_resources.0.desired_vcpus"); ok {
						computeResourceUpdate.DesiredvCpus = aws.Int32(int32(desiredvCpus.(int)))
					} else {
						computeResourceUpdate.DesiredvCpus = aws.Int32(0)
					}
				}

				if d.HasChange("compute_resources.0.min_vcpus") {
					if minVcpus, ok := d.GetOk("compute_resources.0.min_vcpus"); ok {
						computeResourceUpdate.MinvCpus = aws.Int32(int32(minVcpus.(int)))
					} else {
						computeResourceUpdate.MinvCpus = aws.Int32(0)
					}
				}

				if d.HasChange("compute_resources.0.bid_percentage") {
					if bidPercentage, ok := d.GetOk("compute_resources.0.bid_percentage"); ok {
						computeResourceUpdate.BidPercentage = aws.Int32(int32(bidPercentage.(int)))
					} else {
						computeResourceUpdate.BidPercentage = aws.Int32(0)
					}
				}

				if d.HasChange("compute_resources.0.ec2_configuration") {
					defaultImageType := "ECS_AL2"
					if _, ok := d.GetOk("eks_configuration.#"); ok {
						defaultImageType = "EKS_AL2"
					}
					ec2Configuration := d.Get("compute_resources.0.ec2_configuration").([]any)
					computeResourceUpdate.Ec2Configuration = expandEC2ConfigurationsUpdate(ec2Configuration, defaultImageType)
				}

				if d.HasChange("compute_resources.0.ec2_key_pair") {
					if keyPair, ok := d.GetOk("compute_resources.0.ec2_key_pair"); ok {
						computeResourceUpdate.Ec2KeyPair = aws.String(keyPair.(string))
					} else {
						computeResourceUpdate.Ec2KeyPair = aws.String("")
					}
				}

				if d.HasChange("compute_resources.0.image_id") {
					if imageId, ok := d.GetOk("compute_resources.0.image_id"); ok {
						computeResourceUpdate.ImageId = aws.String(imageId.(string))
					} else {
						computeResourceUpdate.ImageId = aws.String("")
					}
				}

				if d.HasChange("compute_resources.0.instance_role") {
					if instanceRole, ok := d.GetOk("compute_resources.0.instance_role"); ok {
						computeResourceUpdate.InstanceRole = aws.String(instanceRole.(string))
					} else {
						computeResourceUpdate.InstanceRole = aws.String("")
					}
				}

				if d.HasChange("compute_resources.0.instance_type") {
					computeResourceUpdate.InstanceTypes = flex.ExpandStringValueSet(d.Get("compute_resources.0.instance_type").(*schema.Set))
				}

				if d.HasChange("compute_resources.0.launch_template") {
					launchTemplate := d.Get("compute_resources.0.launch_template").([]any)
					computeResourceUpdate.LaunchTemplate = expandLaunchTemplateSpecificationUpdate(launchTemplate)
				}

				if d.HasChange("compute_resources.0.tags") {
					if tags, ok := d.GetOk("compute_resources.0.tags"); ok {
						computeResourceUpdate.Tags = svcTags(tftags.New(ctx, tags.(map[string]any)).IgnoreAWS())
					} else {
						computeResourceUpdate.Tags = map[string]string{}
					}
				}
			}

			input.ComputeResources = computeResourceUpdate
		}

		_, err := conn.UpdateComputeEnvironment(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Compute Environment (%s): %s", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceComputeEnvironmentRead(ctx, d, meta)...)
}

func resourceComputeEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	log.Printf("[DEBUG] Disabling Batch Compute Environment: %s", d.Id())
	updateInput := batch.UpdateComputeEnvironmentInput{
		ComputeEnvironment: aws.String(d.Id()),
		State:              awstypes.CEStateDisabled,
	}
	_, err := conn.UpdateComputeEnvironment(ctx, &updateInput)

	if errs.IsAErrorMessageContains[*awstypes.ClientException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Batch Compute Environment (%s): %s", d.Id(), err)
	}

	if _, err := waitComputeEnvironmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		log.Printf("[WARN] error waiting for Batch Compute Environment (%s) disable: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Batch Compute Environment: %s", d.Id())
	deleteInput := batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(d.Id()),
	}
	_, err = conn.DeleteComputeEnvironment(ctx, &deleteInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Batch Compute Environment (%s): %s", d.Id(), err)
	}

	if _, err := waitComputeEnvironmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceComputeEnvironmentCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	if computeEnvironmentType := strings.ToUpper(diff.Get(names.AttrType).(string)); computeEnvironmentType == string(awstypes.CETypeUnmanaged) {
		// UNMANAGED compute environments can have no compute_resources configured.
		if v, ok := diff.GetOk("compute_resources"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			return fmt.Errorf("no `compute_resources` can be specified when `type` is %q", computeEnvironmentType)
		}
	}

	if diff.Id() != "" {
		// Update.

		fargateComputeResources := isFargateType(awstypes.CRType(diff.Get("compute_resources.0.type").(string)))

		if !isUpdatableComputeEnvironment(diff) {
			if diff.HasChange("compute_resources.0.security_group_ids") && !fargateComputeResources {
				if err := diff.ForceNew("compute_resources.0.security_group_ids"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.subnets") && !fargateComputeResources {
				if err := diff.ForceNew("compute_resources.0.subnets"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.allocation_strategy") {
				if err := diff.ForceNew("compute_resources.0.allocation_strategy"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.bid_percentage") {
				if err := diff.ForceNew("compute_resources.0.bid_percentage"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_configuration.#") {
				if err := diff.ForceNew("compute_resources.0.ec2_configuration.#"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_configuration.0.image_id_override") {
				if err := diff.ForceNew("compute_resources.0.ec2_configuration.0.image_id_override"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_configuration.0.image_type") {
				if err := diff.ForceNew("compute_resources.0.ec2_configuration.0.image_type"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_key_pair") {
				if err := diff.ForceNew("compute_resources.0.ec2_key_pair"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.image_id") {
				if err := diff.ForceNew("compute_resources.0.image_id"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.instance_role") {
				if err := diff.ForceNew("compute_resources.0.instance_role"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.instance_type") {
				if err := diff.ForceNew("compute_resources.0.instance_type"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.#") {
				if err := diff.ForceNew("compute_resources.0.launch_template.#"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.0.launch_template_id") {
				if err := diff.ForceNew("compute_resources.0.launch_template.0.launch_template_id"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.0.launch_template_name") {
				if err := diff.ForceNew("compute_resources.0.launch_template.0.launch_template_name"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.0.version") {
				if err := diff.ForceNew("compute_resources.0.launch_template.0.version"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.tags") {
				if err := diff.ForceNew("compute_resources.0.tags"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func findComputeEnvironmentDetailByName(ctx context.Context, conn *batch.Client, name string) (*awstypes.ComputeEnvironmentDetail, error) {
	input := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []string{name},
	}

	output, err := findComputeEnvironmentDetail(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.CEStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findComputeEnvironmentDetail(ctx context.Context, conn *batch.Client, input *batch.DescribeComputeEnvironmentsInput) (*awstypes.ComputeEnvironmentDetail, error) {
	output, err := findComputeEnvironmentDetails(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findComputeEnvironmentDetails(ctx context.Context, conn *batch.Client, input *batch.DescribeComputeEnvironmentsInput) ([]awstypes.ComputeEnvironmentDetail, error) {
	var output []awstypes.ComputeEnvironmentDetail

	pages := batch.NewDescribeComputeEnvironmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ComputeEnvironments...)
	}

	return output, nil
}

func statusComputeEnvironment(ctx context.Context, conn *batch.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findComputeEnvironmentDetailByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitComputeEnvironmentCreated(ctx context.Context, conn *batch.Client, name string, timeout time.Duration) (*awstypes.ComputeEnvironmentDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CEStatusCreating),
		Target:  enum.Slice(awstypes.CEStatusValid),
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ComputeEnvironmentDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentUpdated(ctx context.Context, conn *batch.Client, name string, timeout time.Duration) (*awstypes.ComputeEnvironmentDetail, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CEStatusUpdating),
		Target:  enum.Slice(awstypes.CEStatusValid),
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ComputeEnvironmentDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentDeleted(ctx context.Context, conn *batch.Client, name string, timeout time.Duration) (*awstypes.ComputeEnvironmentDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CEStatusDeleting),
		Target:  []string{},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ComputeEnvironmentDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func isFargateType(computeResourceType awstypes.CRType) bool {
	if computeResourceType == awstypes.CRTypeFargate || computeResourceType == awstypes.CRTypeFargateSpot {
		return true
	}
	return false
}

func isUpdatableComputeEnvironment(diff *schema.ResourceDiff) bool {
	if !isServiceLinkedRoleDiff(diff) {
		return false
	}
	if !isUpdatableAllocationStrategyDiff(diff) {
		return false
	}
	return true
}

func isServiceLinkedRoleDiff(diff *schema.ResourceDiff) bool {
	var before, after string
	if diff.HasChange(names.AttrServiceRole) {
		beforeRaw, afterRaw := diff.GetChange(names.AttrServiceRole)
		before, _ = beforeRaw.(string)
		after, _ := afterRaw.(string)
		return isServiceLinkedRole(before) && isServiceLinkedRole(after)
	}
	afterRaw, _ := diff.GetOk(names.AttrServiceRole)
	after, _ = afterRaw.(string)
	return isServiceLinkedRole(after)
}

func isServiceLinkedRole(roleArn string) bool {
	if roleArn == "" {
		// Empty role ARN defaults to AWS service-linked role
		return true
	}
	re := regexache.MustCompile(`arn:[^:]+:iam::\d{12}:role/aws-service-role/batch\.amazonaws\.com/*`)
	return re.MatchString(roleArn)
}

func isUpdatableAllocationStrategyDiff(diff *schema.ResourceDiff) bool {
	var before, after string
	if computeResourcesCount, ok := diff.Get("compute_resources.#").(int); ok {
		if computeResourcesCount > 0 {
			if diff.HasChange("compute_resources.0.allocation_strategy") {
				beforeRaw, afterRaw := diff.GetChange("compute_resources.0.allocation_strategy")
				before, _ = beforeRaw.(string)
				after, _ = afterRaw.(string)
				return isUpdatableAllocationStrategy(awstypes.CRAllocationStrategy(before)) && isUpdatableAllocationStrategy(awstypes.CRAllocationStrategy(after))
			}
			afterRaw, _ := diff.GetOk("compute_resources.0.allocation_strategy")
			after, _ := afterRaw.(string)
			return isUpdatableAllocationStrategy(awstypes.CRAllocationStrategy(after))
		}
	}
	return false
}

func isUpdatableAllocationStrategy(allocationStrategy awstypes.CRAllocationStrategy) bool {
	return allocationStrategy == awstypes.CRAllocationStrategyBestFitProgressive || allocationStrategy == awstypes.CRAllocationStrategySpotCapacityOptimized
}

func expandComputeResource(ctx context.Context, tfMap map[string]any) *awstypes.ComputeResource {
	if tfMap == nil {
		return nil
	}

	var computeResourceType string

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		computeResourceType = v
	}

	apiObject := &awstypes.ComputeResource{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = awstypes.CRAllocationStrategy(v)
	}

	if v, ok := tfMap["bid_percentage"].(int); ok && v != 0 {
		apiObject.BidPercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["desired_vcpus"].(int); ok && v != 0 {
		apiObject.DesiredvCpus = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ec2_configuration"].([]any); ok && len(v) > 0 {
		apiObject.Ec2Configuration = expandEC2Configurations(v)
	}

	if v, ok := tfMap["ec2_key_pair"].(string); ok && v != "" {
		apiObject.Ec2KeyPair = aws.String(v)
	}

	if v, ok := tfMap["image_id"].(string); ok && v != "" {
		apiObject.ImageId = aws.String(v)
	}

	if v, ok := tfMap["instance_role"].(string); ok && v != "" {
		apiObject.InstanceRole = aws.String(v)
	}

	if v, ok := tfMap[names.AttrInstanceType].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrLaunchTemplate].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchTemplate = expandLaunchTemplateSpecification(v[0].(map[string]any))
	}

	if v, ok := tfMap["max_vcpus"].(int); ok && v != 0 {
		apiObject.MaxvCpus = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_vcpus"].(int); ok && v != 0 {
		apiObject.MinvCpus = aws.Int32(int32(v))
	} else if computeResourceType := strings.ToUpper(computeResourceType); computeResourceType == string(awstypes.CRTypeEc2) || computeResourceType == string(awstypes.CRTypeSpot) {
		apiObject.MinvCpus = aws.Int32(0)
	}

	if v, ok := tfMap["placement_group"].(string); ok && v != "" {
		apiObject.PlacementGroup = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["spot_iam_fleet_role"].(string); ok && v != "" {
		apiObject.SpotIamFleetRole = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrTags].(map[string]any); ok && len(v) > 0 {
		apiObject.Tags = svcTags(tftags.New(ctx, v).IgnoreAWS())
	}

	if computeResourceType != "" {
		apiObject.Type = awstypes.CRType(computeResourceType)
	}

	return apiObject
}

func expandEKSConfiguration(tfMap map[string]any) *awstypes.EksConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EksConfiguration{}

	if v, ok := tfMap["eks_cluster_arn"].(string); ok && v != "" {
		apiObject.EksClusterArn = aws.String(v)
	}

	if v, ok := tfMap["kubernetes_namespace"].(string); ok && v != "" {
		apiObject.KubernetesNamespace = aws.String(v)
	}

	return apiObject
}

func expandEC2Configuration(tfMap map[string]any) *awstypes.Ec2Configuration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Ec2Configuration{}

	if v, ok := tfMap["image_id_override"].(string); ok && v != "" {
		apiObject.ImageIdOverride = aws.String(v)
	}

	if v, ok := tfMap["image_type"].(string); ok && v != "" {
		apiObject.ImageType = aws.String(v)
	}

	return apiObject
}

func expandEC2Configurations(tfList []any) []awstypes.Ec2Configuration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Ec2Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandEC2Configuration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateSpecification(tfMap map[string]any) *awstypes.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateSpecification{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandEC2ConfigurationsUpdate(tfList []any, defaultImageType string) []awstypes.Ec2Configuration {
	if len(tfList) == 0 {
		return []awstypes.Ec2Configuration{
			{
				ImageType: aws.String(defaultImageType),
			},
		}
	}

	var apiObjects []awstypes.Ec2Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandEC2Configuration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateSpecificationUpdate(tfList []any) *awstypes.LaunchTemplateSpecification {
	if len(tfList) == 0 || tfList[0] == nil {
		// delete any existing launch template configuration
		return &awstypes.LaunchTemplateSpecification{
			LaunchTemplateId: aws.String(""),
		}
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.LaunchTemplateSpecification{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok {
		apiObject.Version = aws.String(v)
	} else {
		apiObject.Version = aws.String("")
	}

	return apiObject
}

func flattenComputeResource(ctx context.Context, apiObject *awstypes.ComputeResource) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"allocation_strategy": apiObject.AllocationStrategy,
		names.AttrType:        apiObject.Type,
	}

	if v := apiObject.BidPercentage; v != nil {
		tfMap["bid_percentage"] = aws.ToInt32(v)
	}

	if v := apiObject.DesiredvCpus; v != nil {
		tfMap["desired_vcpus"] = aws.ToInt32(v)
	}

	if v := apiObject.Ec2Configuration; v != nil {
		tfMap["ec2_configuration"] = flattenEC2Configurations(v)
	}

	if v := apiObject.Ec2KeyPair; v != nil {
		tfMap["ec2_key_pair"] = aws.ToString(v)
	}

	if v := apiObject.ImageId; v != nil {
		tfMap["image_id"] = aws.ToString(v)
	}

	if v := apiObject.InstanceRole; v != nil {
		tfMap["instance_role"] = aws.ToString(v)
	}

	if v := apiObject.InstanceTypes; v != nil {
		tfMap[names.AttrInstanceType] = v
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap[names.AttrLaunchTemplate] = []any{flattenLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxvCpus; v != nil {
		tfMap["max_vcpus"] = aws.ToInt32(v)
	}

	if v := apiObject.MinvCpus; v != nil {
		tfMap["min_vcpus"] = aws.ToInt32(v)
	}

	if v := apiObject.PlacementGroup; v != nil {
		tfMap["placement_group"] = aws.ToString(v)
	}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap[names.AttrSecurityGroupIDs] = v
	}

	if v := apiObject.SpotIamFleetRole; v != nil {
		tfMap["spot_iam_fleet_role"] = aws.ToString(v)
	}

	if v := apiObject.Subnets; v != nil {
		tfMap[names.AttrSubnets] = v
	}

	if v := apiObject.Tags; v != nil {
		tfMap[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return tfMap
}

func flattenEKSConfiguration(apiObject *awstypes.EksConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.EksClusterArn; v != nil {
		tfMap["eks_cluster_arn"] = aws.ToString(v)
	}

	if v := apiObject.KubernetesNamespace; v != nil {
		tfMap["kubernetes_namespace"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEC2Configuration(apiObject *awstypes.Ec2Configuration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ImageIdOverride; v != nil {
		tfMap["image_id_override"] = aws.ToString(v)
	}

	if v := apiObject.ImageType; v != nil {
		tfMap["image_type"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEC2Configurations(apiObjects []awstypes.Ec2Configuration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEC2Configuration(&apiObject))
	}

	return tfList
}

func flattenLaunchTemplateSpecification(apiObject *awstypes.LaunchTemplateSpecification) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.ToString(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.ToString(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap[names.AttrVersion] = aws.ToString(v)
	}

	return tfMap
}

func expandComputeEnvironmentUpdatePolicy(tfList []any) *awstypes.UpdatePolicy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.UpdatePolicy{
		JobExecutionTimeoutMinutes: aws.Int64(int64(tfMap["job_execution_timeout_minutes"].(int))),
		TerminateJobsOnUpdate:      aws.Bool(tfMap["terminate_jobs_on_update"].(bool)),
	}

	return apiObject
}

func flattenComputeEnvironmentUpdatePolicy(apiObject *awstypes.UpdatePolicy) []any {
	if apiObject == nil {
		return []any{}
	}

	m := map[string]any{
		"job_execution_timeout_minutes": aws.ToInt64(apiObject.JobExecutionTimeoutMinutes),
		"terminate_jobs_on_update":      aws.ToBool(apiObject.TerminateJobsOnUpdate),
	}

	return []any{m}
}
