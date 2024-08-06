// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appautoscaling_target", name="Target")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types;awstypes;awstypes.ScalableTarget")
// @Testing(importStateIdFunc="testAccTargetImportStateIdFunc")
// @Testing(skipEmptyTags=true)
func resourceTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetCreate,
		ReadWithoutTimeout:   resourceTargetRead,
		UpdateWithoutTimeout: resourceTargetUpdate,
		DeleteWithoutTimeout: resourceTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTargetImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrMaxCapacity: {
				Type:     schema.TypeInt,
				Required: true,
			},
			"min_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"scalable_dimension": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	resourceID := d.Get(names.AttrResourceID).(string)
	input := &applicationautoscaling.RegisterScalableTargetInput{
		MaxCapacity:       aws.Int32(int32(d.Get(names.AttrMaxCapacity).(int))),
		MinCapacity:       aws.Int32(int32(d.Get("min_capacity").(int))),
		ResourceId:        aws.String(resourceID),
		ScalableDimension: awstypes.ScalableDimension(d.Get("scalable_dimension").(string)),
		ServiceNamespace:  awstypes.ServiceNamespace(d.Get("service_namespace").(string)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleARN = aws.String(v.(string))
	}

	err := registerScalableTarget(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Application AutoScaling Target (%s): %s", resourceID, err)
	}

	d.SetId(resourceID)

	return append(diags, resourceTargetRead(ctx, d, meta)...)
}

func resourceTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, 2*time.Minute,
		func() (interface{}, error) {
			return FindTargetByThreePartKey(ctx, conn, d.Id(), d.Get("service_namespace").(string), d.Get("scalable_dimension").(string))
		},
		d.IsNewResource(),
	)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Application AutoScaling Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Application AutoScaling Target (%s): %s", d.Id(), err)
	}

	t := outputRaw.(*awstypes.ScalableTarget)

	d.Set(names.AttrARN, t.ScalableTargetARN)
	d.Set(names.AttrMaxCapacity, t.MaxCapacity)
	d.Set("min_capacity", t.MinCapacity)
	d.Set(names.AttrResourceID, t.ResourceId)
	d.Set(names.AttrRoleARN, t.RoleARN)
	d.Set("scalable_dimension", t.ScalableDimension)
	d.Set("service_namespace", t.ServiceNamespace)

	return diags
}

func resourceTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &applicationautoscaling.RegisterScalableTargetInput{
			MaxCapacity:       aws.Int32(int32(d.Get(names.AttrMaxCapacity).(int))),
			MinCapacity:       aws.Int32(int32(d.Get("min_capacity").(int))),
			ResourceId:        aws.String(d.Id()),
			ScalableDimension: awstypes.ScalableDimension(d.Get("scalable_dimension").(string)),
			ServiceNamespace:  awstypes.ServiceNamespace(d.Get("service_namespace").(string)),
		}

		if v, ok := d.GetOk(names.AttrRoleARN); ok {
			input.RoleARN = aws.String(v.(string))
		}

		err := registerScalableTarget(ctx, conn, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Application AutoScaling Target (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTargetRead(ctx, d, meta)...)
}

func resourceTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	input := &applicationautoscaling.DeregisterScalableTargetInput{
		ResourceId:        aws.String(d.Id()),
		ScalableDimension: awstypes.ScalableDimension(d.Get("scalable_dimension").(string)),
		ServiceNamespace:  awstypes.ServiceNamespace(d.Get("service_namespace").(string)),
	}

	log.Printf("[INFO] Deleting Application AutoScaling Target: %s", d.Id())
	_, err := conn.DeregisterScalableTarget(ctx, input)

	if errs.IsA[*awstypes.ObjectNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Application AutoScaling Target (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 5*time.Minute, func() (interface{}, error) {
		return FindTargetByThreePartKey(ctx, conn, d.Id(), d.Get("service_namespace").(string), d.Get("scalable_dimension").(string))
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Application AutoScaling Target (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindTargetByThreePartKey(ctx context.Context, conn *applicationautoscaling.Client, resourceID, namespace, dimension string) (*awstypes.ScalableTarget, error) {
	input := &applicationautoscaling.DescribeScalableTargetsInput{
		ResourceIds:       []string{resourceID},
		ScalableDimension: awstypes.ScalableDimension(dimension),
		ServiceNamespace:  awstypes.ServiceNamespace(namespace),
	}
	var output []awstypes.ScalableTarget

	pages := applicationautoscaling.NewDescribeScalableTargetsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ScalableTargets...)
	}

	target, err := tfresource.AssertSingleValueResult(output)

	if err != nil {
		return nil, err
	}

	if aws.ToString(target.ResourceId) != resourceID || string(target.ScalableDimension) != dimension || string(target.ServiceNamespace) != namespace {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return target, nil
}

func resourceTargetImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")

	if len(idParts) < 3 {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-namespace>/<resource-id>/<scalable-dimension>", d.Id())
	}

	serviceNamespace := idParts[0]
	resourceId := strings.Join(idParts[1:len(idParts)-1], "/")
	scalableDimension := idParts[len(idParts)-1]

	if serviceNamespace == "" || resourceId == "" || scalableDimension == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-namespace>/<resource-id>/<scalable-dimension>", d.Id())
	}

	d.Set("service_namespace", serviceNamespace)
	d.Set(names.AttrResourceID, resourceId)
	d.Set("scalable_dimension", scalableDimension)
	d.SetId(resourceId)

	return []*schema.ResourceData{d}, nil
}

func registerScalableTarget(ctx context.Context, conn *applicationautoscaling.Client, input *applicationautoscaling.RegisterScalableTargetInput) error {
	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.RegisterScalableTarget(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Unable to assume IAM role") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "ECS service doesn't exist") {
				return true, err
			}

			return false, err
		},
	)

	return err
}
