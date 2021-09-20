package transfer

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceServer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"identity_provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"invocation_role": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"logging_role": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"protocols": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"security_policy_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Read: dataSourceServerRead,
	}
}

func dataSourceServerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn

	serverID := d.Get("server_id").(string)

	output, err := FindServerByID(conn, serverID)

	if err != nil {
		return fmt.Errorf("error reading Transfer Server (%s): %w", serverID, err)
	}

	d.SetId(aws.StringValue(output.ServerId))
	d.Set("arn", output.Arn)
	d.Set("certificate", output.Certificate)
	d.Set("domain", output.Domain)
	d.Set("endpoint", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.server.transfer", serverID)))
	d.Set("endpoint_type", output.EndpointType)
	d.Set("identity_provider_type", output.IdentityProviderType)
	if output.IdentityProviderDetails != nil {
		d.Set("invocation_role", output.IdentityProviderDetails.InvocationRole)
	} else {
		d.Set("invocation_role", "")
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("protocols", aws.StringValueSlice(output.Protocols))
	d.Set("security_policy_name", output.SecurityPolicyName)
	if output.IdentityProviderDetails != nil {
		d.Set("url", output.IdentityProviderDetails.Url)
	} else {
		d.Set("url", "")
	}

	return nil
}
