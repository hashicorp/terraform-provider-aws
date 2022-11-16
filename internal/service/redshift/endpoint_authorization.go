package redshift

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEndpointAuthorization() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointAuthorizationCreate,
		Read:   resourceEndpointAuthorizationRead,
		Update: resourceEndpointAuthorizationUpdate,
		Delete: resourceEndpointAuthorizationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"allowed_all_vpcs": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"endpoint_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"grantee": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grantor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceEndpointAuthorizationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	account := d.Get("account").(string)
	input := redshift.AuthorizeEndpointAccessInput{
		Account:           aws.String(account),
		ClusterIdentifier: aws.String(d.Get("cluster_identifier").(string)),
	}

	if v, ok := d.GetOk("vpc_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.VpcIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.AuthorizeEndpointAccess(&input)
	if err != nil {
		return fmt.Errorf("creating Redshift Endpoint Authorization: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", account, aws.StringValue(output.ClusterIdentifier)))
	log.Printf("[INFO] Redshift Endpoint Authorization ID: %s", d.Id())

	return resourceEndpointAuthorizationRead(d, meta)
}

func resourceEndpointAuthorizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	endpoint, err := FindEndpointAuthorizationById(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Endpoint Authorization (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Redshift Endpoint Authorization (%s): %w", d.Id(), err)
	}

	d.Set("account", endpoint.Grantee)
	d.Set("grantee", endpoint.Grantee)
	d.Set("grantor", endpoint.Grantor)
	d.Set("cluster_identifier", endpoint.ClusterIdentifier)
	d.Set("vpc_ids", flex.FlattenStringSet(endpoint.AllowedVPCs))
	d.Set("allowed_all_vpcs", endpoint.AllowedAllVPCs)
	d.Set("endpoint_count", endpoint.EndpointCount)

	return nil
}

func resourceEndpointAuthorizationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChanges("vpc_ids") {
		account, clusterId, err := DecodeEndpointAuthorizationID(d.Id())
		if err != nil {
			return err
		}

		o, n := d.GetChange("vpc_ids")
		ns := n.(*schema.Set)
		os := o.(*schema.Set)
		added := ns.Difference(os)
		removed := os.Difference(ns)

		if added.Len() > 0 {
			_, err := conn.AuthorizeEndpointAccess(&redshift.AuthorizeEndpointAccessInput{
				Account:           aws.String(account),
				ClusterIdentifier: aws.String(clusterId),
				VpcIds:            flex.ExpandStringSet(added),
			})

			if err != nil {
				return fmt.Errorf("authorizing Redshift Endpoint Authorization VPCs (%s): %w", d.Id(), err)
			}
		}

		if removed.Len() > 0 {
			_, err := conn.RevokeEndpointAccess(&redshift.RevokeEndpointAccessInput{
				Account:           aws.String(account),
				ClusterIdentifier: aws.String(clusterId),
				VpcIds:            flex.ExpandStringSet(removed),
			})

			if err != nil {
				return fmt.Errorf("revoking Redshift Endpoint Authorization VPCs (%s): %w", d.Id(), err)
			}
		}
	}

	return resourceEndpointAuthorizationRead(d, meta)
}

func resourceEndpointAuthorizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	account, clusterId, err := DecodeEndpointAuthorizationID(d.Id())
	if err != nil {
		return err
	}

	input := &redshift.RevokeEndpointAccessInput{
		Account:           aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
		Force:             aws.Bool(d.Get("force_delete").(bool)),
	}

	_, err = conn.RevokeEndpointAccess(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeEndpointAuthorizationNotFoundFault) {
			return nil
		}
		return err
	}

	return nil
}

func DecodeEndpointAuthorizationID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID to be the form account:cluster_identifier, given: %s", id)
	}

	return idParts[0], idParts[1], nil
}
