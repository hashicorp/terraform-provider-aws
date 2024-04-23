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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_compute_environment", name="Compute Environment")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/batch/types.ComputeEnvironmentDetail")
func ResourceComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceComputeEnvironmentCreate,
		ReadWithoutTimeout:   resourceComputeEnvironmentRead,
		UpdateWithoutTimeout: resourceComputeEnvironmentUpdate,
		DeleteWithoutTimeout: resourceComputeEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceComputeEnvironmentCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
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
							Type:     schema.TypeString,
							Optional: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateDiagFunc: enum.Validate[awstypes.CRAllocationStrategy](),
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
						"instance_type": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"launch_template": {
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
									"version": {
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
						"security_group_ids": {
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
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tags": tftags.TagsSchema(),
						"type": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateDiagFunc: enum.Validate[awstypes.CRType](),
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
			"service_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateDiagFunc: enum.Validate[awstypes.CEState](),
				Default:          awstypes.CEStateEnabled,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateDiagFunc: enum.Validate[awstypes.CEType](),
			},
			"update_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_execution_timeout_minutes": {
							Type:     schema.TypeInt,
							Required: true,
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

func resourceComputeEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	computeEnvironmentName := create.Name(d.Get("compute_environment_name").(string), d.Get("compute_environment_name_prefix").(string))
	computeEnvironmentType := d.Get("type").(string)
	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(computeEnvironmentName),
		ServiceRole:            aws.String(d.Get("service_role").(string)),
		Tags:                   getTagsIn(ctx),
		Type:                   awstypes.CEType(computeEnvironmentType),
	}

	if v, ok := d.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ComputeResources = expandComputeResource(ctx, v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("eks_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EksConfiguration = expandEKSConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("state"); ok {
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
	if v, ok := d.GetOk("update_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		inputUpdateOnCreate := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
			UpdatePolicy:       expandComputeEnvironmentUpdatePolicy(v.([]interface{})),
		}
		log.Printf("[DEBUG] Creating Batch Compute Environment extra arguments: %+v", inputUpdateOnCreate)

		if _, err := conn.UpdateComputeEnvironment(ctx, inputUpdateOnCreate); err != nil {
			return sdkdiag.AppendErrorf(diags, "Create Batch Compute Environment extra arguments through UpdateComputeEnvironment (%s): %s", d.Id(), err)
		}

		if err := waitComputeEnvironmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "Create waiting for Batch Compute Environment (%s) extra arguments through UpdateComputeEnvironment: %s", d.Id(), err)
		}
	}

	return append(diags, resourceComputeEnvironmentRead(ctx, d, meta)...)
}

func resourceComputeEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("compute_environment_name_prefix", create.NamePrefixFromName(aws.ToString(computeEnvironment.ComputeEnvironmentName)))
	if computeEnvironment.ComputeResources != nil {
		if err := d.Set("compute_resources", []interface{}{flattenComputeResource(ctx, computeEnvironment.ComputeResources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting compute_resources: %s", err)
		}
	} else {
		d.Set("compute_resources", nil)
	}
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	if computeEnvironment.EksConfiguration != nil {
		if err := d.Set("eks_configuration", []interface{}{flattenEKSConfiguration(computeEnvironment.EksConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting eks_configuration: %s", err)
		}
	} else {
		d.Set("eks_configuration", nil)
	}
	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("state", computeEnvironment.State)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)
	d.Set("type", string(computeEnvironment.Type))

	if err := d.Set("update_policy", flattenComputeEnvironmentUpdatePolicy(computeEnvironment.UpdatePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting update_policy: %s", err)
	}

	setTagsOut(ctx, computeEnvironment.Tags)

	return diags
}

func resourceComputeEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if d.HasChange("service_role") {
			input.ServiceRole = aws.String(d.Get("service_role").(string))
		}

		if d.HasChange("state") {
			input.State = awstypes.CEState(d.Get("state").(string))
		}

		if d.HasChange("update_policy") {
			input.UpdatePolicy = expandComputeEnvironmentUpdatePolicy(d.Get("update_policy").([]interface{}))
		}

		if computeEnvironmentType := strings.ToUpper(d.Get("type").(string)); computeEnvironmentType == string(awstypes.CETypeManaged) {
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
					computeResourceUpdate.AllocationStrategy = awstypes.CRUpdateAllocationStrategy("")
				}
			}

			computeResourceEnvironmentType := d.Get("compute_resources.0.type").(string)

			if d.HasChange("compute_resources.0.type") {
				computeResourceUpdate.Type = awstypes.CRType(computeResourceEnvironmentType)
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
					ec2Configuration := d.Get("compute_resources.0.ec2_configuration").([]interface{})
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
					launchTemplate := d.Get("compute_resources.0.launch_template").([]interface{})
					computeResourceUpdate.LaunchTemplate = expandLaunchTemplateSpecificationUpdate(launchTemplate)
				}

				if d.HasChange("compute_resources.0.tags") {
					if tags, ok := d.GetOk("compute_resources.0.tags"); ok {
						computeResourceUpdate.Tags = Tags(tftags.New(ctx, tags.(map[string]interface{})).IgnoreAWS())
					} else {
						computeResourceUpdate.Tags = map[string]string{}
					}
				}
			}

			input.ComputeResources = computeResourceUpdate
		}

		log.Printf("[DEBUG] Updating Batch Compute Environment: %+v", input)
		if _, err := conn.UpdateComputeEnvironment(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Compute Environment (%s): %s", d.Id(), err)
		}

		if err := waitComputeEnvironmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceComputeEnvironmentRead(ctx, d, meta)...)
}

func resourceComputeEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	log.Printf("[DEBUG] Disabling Batch Compute Environment: %s", d.Id())
	{
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
			State:              awstypes.CEStateDisabled,
		}

		if _, err := conn.UpdateComputeEnvironment(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling Batch Compute Environment (%s): %s", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentDisabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			log.Printf("[WARN] error waiting for Batch Compute Environment (%s) disable: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Batch Compute Environment: %s", d.Id())
	{
		input := &batch.DeleteComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if _, err := conn.DeleteComputeEnvironment(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Batch Compute Environment (%s): %s", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Batch Compute Environment (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceComputeEnvironmentCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if computeEnvironmentType := strings.ToUpper(diff.Get("type").(string)); computeEnvironmentType == string(awstypes.CETypeUnmanaged) {
		// UNMANAGED compute environments can have no compute_resources configured.
		if v, ok := diff.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			return fmt.Errorf("no `compute_resources` can be specified when `type` is %q", computeEnvironmentType)
		}
	}

	if diff.Id() != "" {
		// Update.

		fargateComputeResources := isFargateType(diff.Get("compute_resources.0.type").(string))

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

	if output.Status == awstypes.CEStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.Status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findComputeEnvironmentDetail(ctx context.Context, conn *batch.Client, input *batch.DescribeComputeEnvironmentsInput) (*awstypes.ComputeEnvironmentDetail, error) {
	output, err := conn.DescribeComputeEnvironments(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.ComputeEnvironments)
}

func statusComputeEnvironment(ctx context.Context, conn *batch.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		computeEnvironmentDetail, err := findComputeEnvironmentDetailByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return computeEnvironmentDetail, string(computeEnvironmentDetail.Status), nil
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
		if output.Status == awstypes.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

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
		if output.Status == awstypes.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentDisabled(ctx context.Context, conn *batch.Client, name string, timeout time.Duration) (*awstypes.ComputeEnvironmentDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CEStatusUpdating),
		Target:  enum.Slice(awstypes.CEStatusValid),
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ComputeEnvironmentDetail); ok {
		if output.Status == awstypes.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentUpdated(ctx context.Context, conn *batch.Client, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CEStatusUpdating),
		Target:  enum.Slice(awstypes.CEStatusValid),
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if _, ok := outputRaw.(*awstypes.ComputeEnvironmentDetail); ok {
		return err
	}

	return err
}

func isFargateType(computeResourceType string) bool {
	if computeResourceType == string(awstypes.CRTypeFargate) || computeResourceType == string(awstypes.CRTypeFargateSpot) {
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
	if diff.HasChange("service_role") {
		beforeRaw, afterRaw := diff.GetChange("service_role")
		before, _ = beforeRaw.(string)
		after, _ := afterRaw.(string)
		return isServiceLinkedRole(before) && isServiceLinkedRole(after)
	}
	afterRaw, _ := diff.GetOk("service_role")
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
				return isUpdatableAllocationStrategy(before) && isUpdatableAllocationStrategy(after)
			}
			afterRaw, _ := diff.GetOk("compute_resources.0.allocation_strategy")
			after, _ := afterRaw.(string)
			return isUpdatableAllocationStrategy(after)
		}
	}
	return false
}

func isUpdatableAllocationStrategy(allocationStrategy string) bool {
	return allocationStrategy == string(awstypes.CRAllocationStrategyBestFitProgressive) || allocationStrategy == string(awstypes.CRAllocationStrategySpotCapacityOptimized)
}

func expandComputeResource(ctx context.Context, tfMap map[string]interface{}) *awstypes.ComputeResource {
	if tfMap == nil {
		return nil
	}

	var computeResourceType string

	if v, ok := tfMap["type"].(string); ok && v != "" {
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

	if v, ok := tfMap["ec2_configuration"].([]interface{}); ok && len(v) > 0 {
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

	if v, ok := tfMap["instance_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchTemplate = expandLaunchTemplateSpecification(v[0].(map[string]interface{}))
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

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["spot_iam_fleet_role"].(string); ok && v != "" {
		apiObject.SpotIamFleetRole = aws.String(v)
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Tags = Tags(tftags.New(ctx, v).IgnoreAWS())
	}

	if computeResourceType != "" {
		apiObject.Type = awstypes.CRType(computeResourceType)
	}

	return apiObject
}

func expandEKSConfiguration(tfMap map[string]interface{}) *awstypes.EksConfiguration {
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

func expandEC2Configuration(tfMap map[string]interface{}) awstypes.Ec2Configuration {
	apiObject := awstypes.Ec2Configuration{}

	if v, ok := tfMap["image_id_override"].(string); ok && v != "" {
		apiObject.ImageIdOverride = aws.String(v)
	}

	if v, ok := tfMap["image_type"].(string); ok && v != "" {
		apiObject.ImageType = aws.String(v)
	}

	return apiObject
}

func expandEC2Configurations(tfList []interface{}) []awstypes.Ec2Configuration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Ec2Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEC2Configuration(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}) *awstypes.LaunchTemplateSpecification {
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

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandEC2ConfigurationsUpdate(tfList []interface{}, defaultImageType string) []awstypes.Ec2Configuration {
	if len(tfList) == 0 {
		return []awstypes.Ec2Configuration{
			{
				ImageType: aws.String(defaultImageType),
			},
		}
	}

	var apiObjects []awstypes.Ec2Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEC2Configuration(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateSpecificationUpdate(tfList []interface{}) *awstypes.LaunchTemplateSpecification {
	if len(tfList) == 0 || tfList[0] == nil {
		// delete any existing launch template configuration
		return &awstypes.LaunchTemplateSpecification{
			LaunchTemplateId: aws.String(""),
		}
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.LaunchTemplateSpecification{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok {
		apiObject.Version = aws.String(v)
	} else {
		apiObject.Version = aws.String("")
	}

	return apiObject
}

func flattenComputeResource(ctx context.Context, apiObject *awstypes.ComputeResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["allocation_strategy"] = string(apiObject.AllocationStrategy)

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
		tfMap["instance_type"] = v
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenLaunchTemplateSpecification(v)}
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
		tfMap["security_group_ids"] = v
	}

	if v := apiObject.SpotIamFleetRole; v != nil {
		tfMap["spot_iam_fleet_role"] = aws.ToString(v)
	}

	if v := apiObject.Subnets; v != nil {
		tfMap["subnets"] = v
	}

	if v := apiObject.Tags; v != nil {
		tfMap["tags"] = KeyValueTags(ctx, v).IgnoreAWS().Map()
	}

	tfMap["type"] = string(apiObject.Type)

	return tfMap
}

func flattenEKSConfiguration(apiObject *awstypes.EksConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EksClusterArn; v != nil {
		tfMap["eks_cluster_arn"] = aws.ToString(v)
	}

	if v := apiObject.KubernetesNamespace; v != nil {
		tfMap["kubernetes_namespace"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEC2Configuration(apiObject awstypes.Ec2Configuration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.ImageIdOverride; v != nil {
		tfMap["image_id_override"] = aws.ToString(v)
	}

	if v := apiObject.ImageType; v != nil {
		tfMap["image_type"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEC2Configurations(apiObjects []awstypes.Ec2Configuration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEC2Configuration(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateSpecification(apiObject *awstypes.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.ToString(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.ToString(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap["version"] = aws.ToString(v)
	}

	return tfMap
}

func expandComputeEnvironmentUpdatePolicy(l []interface{}) *awstypes.UpdatePolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	up := &awstypes.UpdatePolicy{
		JobExecutionTimeoutMinutes: aws.Int64(int64(m["job_execution_timeout_minutes"].(int))),
		TerminateJobsOnUpdate:      aws.Bool(m["terminate_jobs_on_update"].(bool)),
	}

	return up
}

func flattenComputeEnvironmentUpdatePolicy(up *awstypes.UpdatePolicy) []interface{} {
	if up == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"job_execution_timeout_minutes": aws.ToInt64(up.JobExecutionTimeoutMinutes),
		"terminate_jobs_on_update":      aws.ToBool(up.TerminateJobsOnUpdate),
	}

	return []interface{}{m}
}
