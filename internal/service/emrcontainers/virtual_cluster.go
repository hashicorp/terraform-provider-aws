package emrcontainers

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVirtualCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualClusterCreate,
		Read:   resourceVirtualClusterRead,
		Delete: resourceVirtualClusterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[.\-_/#A-Za-z0-9]+`), ""),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVirtualClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRContainersConn

	input := emrcontainers.CreateVirtualClusterInput{
		ContainerProvider: expandEMRContainersContainerProvider(d.Get("container_provider").([]interface{})),
		Name:              aws.String(d.Get("name").(string)),
	}

	log.Printf("[INFO] Creating EMR containers virtual cluster: %s", input)
	out, err := conn.CreateVirtualCluster(&input)
	if err != nil {
		return fmt.Errorf("error creating EMR containers virtual cluster: %w", err)
	}

	d.SetId(aws.StringValue(out.Id))

	return resourceVirtualClusterRead(d, meta)
}

func resourceVirtualClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRContainersConn

	vc, err := FindVirtualClusterByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Containers Virtual Cluster %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EMR Containers Virtual Cluster (%s): %w", d.Id(), err)
	}

	d.Set("arn", vc.Arn)
	if err := d.Set("container_provider", flattenEMRContainersContainerProvider(vc.ContainerProvider)); err != nil {
		return fmt.Errorf("error reading EMR containers virtual cluster (%s): %w", d.Id(), err)
	}
	d.Set("created_at", aws.TimeValue(vc.CreatedAt).String())
	d.Set("name", vc.Name)
	d.Set("state", vc.State)

	return nil
}

func resourceVirtualClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRContainersConn

	log.Printf("[INFO] EMR containers virtual cluster: %s", d.Id())
	_, err := conn.DeleteVirtualCluster(&emrcontainers.DeleteVirtualClusterInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, emrcontainers.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return fmt.Errorf("error deleting EMR containers virtual cluster (%s): %w", d.Id(), err)
	}

	_, err = waitVirtualClusterDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for EMR containers virtual cluster (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func expandEMRContainersContainerProvider(l []interface{}) *emrcontainers.ContainerProvider {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	input := emrcontainers.ContainerProvider{
		Id:   aws.String(m["id"].(string)),
		Type: aws.String(m["type"].(string)),
	}

	if v, ok := m["info"]; ok {
		input.Info = expandEMRContainersContainerInfo(v.([]interface{}))
	}

	return &input
}

func expandEMRContainersContainerInfo(l []interface{}) *emrcontainers.ContainerInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	input := emrcontainers.ContainerInfo{}

	if v, ok := m["eks_info"]; ok {
		input.EksInfo = expandEMRContainersEksInfo(v.([]interface{}))
	}

	return &input
}

func expandEMRContainersEksInfo(l []interface{}) *emrcontainers.EksInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	input := emrcontainers.EksInfo{}

	if v, ok := m["namespace"]; ok {
		input.Namespace = aws.String(v.(string))
	}

	return &input
}

func flattenEMRContainersContainerProvider(cp *emrcontainers.ContainerProvider) []interface{} {
	if cp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	m["id"] = cp.Id
	m["type"] = cp.Type

	if cp.Info != nil {
		m["info"] = flattenEMRContainersContainerInfo(cp.Info)
	}

	return []interface{}{m}
}

func flattenEMRContainersContainerInfo(ci *emrcontainers.ContainerInfo) []interface{} {
	if ci == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if ci.EksInfo != nil {
		m["eks_info"] = flattenEMRContainersEksInfo(ci.EksInfo)
	}

	return []interface{}{m}
}

func flattenEMRContainersEksInfo(ei *emrcontainers.EksInfo) []interface{} {
	if ei == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if ei.Namespace != nil {
		m["namespace"] = ei.Namespace
	}

	return []interface{}{m}
}

func findVirtualCluster(conn *emrcontainers.EMRContainers, input *emrcontainers.DescribeVirtualClusterInput) (*emrcontainers.VirtualCluster, error) {
	output, err := conn.DescribeVirtualCluster(input)

	if tfawserr.ErrCodeEquals(err, emrcontainers.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindVirtualClusterByID(conn *emrcontainers.EMRContainers, id string) (*emrcontainers.VirtualCluster, error) {
	input := &emrcontainers.DescribeVirtualClusterInput{
		Id: aws.String(id),
	}

	output, err := findVirtualCluster(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == emrcontainers.VirtualClusterStateTerminated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusVirtualCluster(conn *emrcontainers.EMRContainers, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVirtualClusterByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitVirtualClusterDeleted(conn *emrcontainers.EMRContainers, id string, timeout time.Duration) (*emrcontainers.VirtualCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{emrcontainers.VirtualClusterStateTerminating},
		Target:  []string{},
		Refresh: statusVirtualCluster(conn, id),
		Timeout: timeout,
		Delay:   1 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*emrcontainers.VirtualCluster); ok {
		return v, err
	}

	return nil, err
}
