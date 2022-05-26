package redshift

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceEndpointAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointAccessCreate,
		Read:   resourceEndpointAccessRead,
		Update: resourceEndpointAccessUpdate,
		Delete: resourceEndpointAccessDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"endpoint_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 30),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
				),
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resource_owner": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"subnet_group_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"vpc_endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_interface": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"network_interface_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"private_ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"vpc_endpoint_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceEndpointAccessCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	createOpts := redshift.CreateEndpointAccessInput{
		EndpointName:    aws.String(d.Get("endpoint_name").(string)),
		SubnetGroupName: aws.String(d.Get("subnet_group_name").(string)),
	}

	if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		createOpts.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("cluster_identifier"); ok {
		createOpts.ClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_owner"); ok {
		createOpts.ResourceOwner = aws.String(v.(string))
	}

	_, err := conn.CreateEndpointAccess(&createOpts)
	if err != nil {
		return fmt.Errorf("error creating Redshift endpoint access: %w", err)
	}

	d.SetId(aws.StringValue(createOpts.EndpointName))
	log.Printf("[INFO] Redshift endpoint access ID: %s", d.Id())

	if _, err := waitEndpointAccessActive(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Redshift Endpoint Access (%s) to be active: %w", d.Id(), err)
	}

	return resourceEndpointAccessRead(d, meta)
}

func resourceEndpointAccessRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	endpoint, err := FindEndpointAccessByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift endpoint access (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift endpoint access (%s): %w", d.Id(), err)
	}

	d.Set("endpoint_name", endpoint.EndpointName)
	d.Set("subnet_group_name", endpoint.SubnetGroupName)
	d.Set("vpc_security_group_ids", vpcSgsIdsToSlice(endpoint.VpcSecurityGroups))
	d.Set("resource_owner", endpoint.ResourceOwner)
	d.Set("cluster_identifier", endpoint.ClusterIdentifier)
	d.Set("port", endpoint.Port)
	d.Set("address", endpoint.Address)

	if err := d.Set("vpc_endpoint", flattenVPCEndpoint(endpoint.VpcEndpoint)); err != nil {
		return fmt.Errorf("setting vpc_endpoint: %w", err)
	}

	return nil
}

func resourceEndpointAccessUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChanges("vpc_security_group_ids") {
		_, n := d.GetChange("vpc_security_group_ids")
		if n == nil {
			n = new(schema.Set)
		}
		ns := n.(*schema.Set)

		var sIds []*string
		for _, s := range ns.List() {
			sIds = append(sIds, aws.String(s.(string)))
		}

		_, err := conn.ModifyEndpointAccess(&redshift.ModifyEndpointAccessInput{
			EndpointName:        aws.String(d.Id()),
			VpcSecurityGroupIds: sIds,
		})

		if err != nil {
			return fmt.Errorf("error updating Redshift endpoint access (%s): %w", d.Id(), err)
		}

		if _, err := waitEndpointAccessActive(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for Redshift Endpoint Access (%s) to be active: %w", d.Id(), err)
		}
	}

	return resourceEndpointAccessRead(d, meta)
}

func resourceEndpointAccessDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	_, err := conn.DeleteEndpointAccess(&redshift.DeleteEndpointAccessInput{
		EndpointName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeEndpointNotFoundFault) {
			return nil
		}
		return err
	}

	if _, err := waitEndpointAccessDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Redshift Endpoint Access (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func vpcSgsIdsToSlice(vpsSgsIds []*redshift.VpcSecurityGroupMembership) []string {
	VpcSgsSlice := make([]string, 0, len(vpsSgsIds))
	for _, s := range vpsSgsIds {
		VpcSgsSlice = append(VpcSgsSlice, *s.VpcSecurityGroupId)
	}
	return VpcSgsSlice
}

func flattenVPCEndpoint(apiObject *redshift.VpcEndpoint) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NetworkInterfaces; v != nil {
		tfMap["network_interface"] = flattenNetworkInterfaces(v)
	}

	if v := apiObject.VpcEndpointId; v != nil {
		tfMap["vpc_endpoint_id"] = aws.StringValue(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenNetworkInterface(apiObject *redshift.NetworkInterface) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap["availability_zone"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	if v := apiObject.PrivateIpAddress; v != nil {
		tfMap["private_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap["subnet_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenNetworkInterfaces(apiObjects []*redshift.NetworkInterface) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterface(apiObject))
	}

	return tfList
}
