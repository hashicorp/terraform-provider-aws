// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_emrcontainers_virtual_cluster", name="Virtual Cluster")
// @Tags(identifierAttribute="arn")
func ResourceVirtualCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualClusterCreate,
		ReadWithoutTimeout:   resourceVirtualClusterRead,
		UpdateWithoutTimeout: resourceVirtualClusterUpdate,
		DeleteWithoutTimeout: resourceVirtualClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						// According to https://docs.aws.amazon.com/emr-on-eks/latest/APIReference/API_ContainerProvider.html
						// The info and the eks_info are optional but the API raises ValidationException without the fields
						"info": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eks_info": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"namespace": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(emrcontainers.ContainerProviderType_Values(), false),
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_./#-]+`), "must contain only alphanumeric, hyphen, underscore, dot and # characters"),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVirtualClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRContainersConn(ctx)

	name := d.Get("name").(string)
	input := &emrcontainers.CreateVirtualClusterInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("container_provider"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ContainerProvider = expandContainerProvider(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateVirtualClusterWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Containers Virtual Cluster (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	return append(diags, resourceVirtualClusterRead(ctx, d, meta)...)
}

func resourceVirtualClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRContainersConn(ctx)

	vc, err := FindVirtualClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Containers Virtual Cluster %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Containers Virtual Cluster (%s): %s", d.Id(), err)
	}

	d.Set("arn", vc.Arn)
	if vc.ContainerProvider != nil {
		if err := d.Set("container_provider", []interface{}{flattenContainerProvider(vc.ContainerProvider)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting container_provider: %s", err)
		}
	} else {
		d.Set("container_provider", nil)
	}
	d.Set("name", vc.Name)

	setTagsOut(ctx, vc.Tags)

	return diags
}

func resourceVirtualClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceVirtualClusterRead(ctx, d, meta)
}

func resourceVirtualClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRContainersConn(ctx)

	log.Printf("[INFO] Deleting EMR Containers Virtual Cluster: %s", d.Id())
	_, err := conn.DeleteVirtualClusterWithContext(ctx, &emrcontainers.DeleteVirtualClusterInput{
		Id: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, emrcontainers.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Containers Virtual Cluster (%s): %s", d.Id(), err)
	}

	if _, err = waitVirtualClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Containers Virtual Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandContainerProvider(tfMap map[string]interface{}) *emrcontainers.ContainerProvider {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.ContainerProvider{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["info"].([]interface{}); ok && len(v) > 0 {
		apiObject.Info = expandContainerInfo(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandContainerInfo(tfMap map[string]interface{}) *emrcontainers.ContainerInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.ContainerInfo{}

	if v, ok := tfMap["eks_info"].([]interface{}); ok && len(v) > 0 {
		apiObject.EksInfo = expandEKSInfo(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandEKSInfo(tfMap map[string]interface{}) *emrcontainers.EksInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrcontainers.EksInfo{}

	if v, ok := tfMap["namespace"].(string); ok && v != "" {
		apiObject.Namespace = aws.String(v)
	}

	return apiObject
}

func flattenContainerProvider(apiObject *emrcontainers.ContainerProvider) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Id; v != nil {
		tfMap["id"] = aws.StringValue(v)
	}

	if v := apiObject.Info; v != nil {
		tfMap["info"] = []interface{}{flattenContainerInfo(v)}
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenContainerInfo(apiObject *emrcontainers.ContainerInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EksInfo; v != nil {
		tfMap["eks_info"] = []interface{}{flattenEKSInfo(v)}
	}

	return tfMap
}

func flattenEKSInfo(apiObject *emrcontainers.EksInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Namespace; v != nil {
		tfMap["namespace"] = aws.StringValue(v)
	}

	return tfMap
}

func findVirtualCluster(ctx context.Context, conn *emrcontainers.EMRContainers, input *emrcontainers.DescribeVirtualClusterInput) (*emrcontainers.VirtualCluster, error) {
	output, err := conn.DescribeVirtualClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, emrcontainers.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VirtualCluster == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VirtualCluster, nil
}

func FindVirtualClusterByID(ctx context.Context, conn *emrcontainers.EMRContainers, id string) (*emrcontainers.VirtualCluster, error) {
	input := &emrcontainers.DescribeVirtualClusterInput{
		Id: aws.String(id),
	}

	output, err := findVirtualCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == emrcontainers.VirtualClusterStateTerminated {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusVirtualCluster(ctx context.Context, conn *emrcontainers.EMRContainers, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVirtualClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitVirtualClusterDeleted(ctx context.Context, conn *emrcontainers.EMRContainers, id string, timeout time.Duration) (*emrcontainers.VirtualCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{emrcontainers.VirtualClusterStateTerminating},
		Target:  []string{},
		Refresh: statusVirtualCluster(ctx, conn, id),
		Timeout: timeout,
		Delay:   1 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*emrcontainers.VirtualCluster); ok {
		return v, err
	}

	return nil, err
}
