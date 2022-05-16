package kms

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceCustomKeyStore() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCustomKeyStoreRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cloudhsm_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trust_anchor_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCustomKeyStoreRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	input := &kms.DescribeCustomKeyStoresInput{}

	if v, ok := d.GetOk("id"); ok {
		input.CustomKeyStoreId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		input.CustomKeyStoreName = aws.String(v.(string))
	}

	output, err := conn.DescribeCustomKeyStores(input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeCustomKeyStoreNotFoundException) {
		return &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return fmt.Errorf("error reading KMS Custom Key Store: %w", err)
	}

	if output == nil || len(output.CustomKeyStores) == 0 || output.CustomKeyStores[0] == nil {
		return tfresource.NewEmptyResultError(input)
	}

	if count := len(output.CustomKeyStores); count > 1 {
		return tfresource.NewTooManyResultsError(count, input)
	}

	keyStore := output.CustomKeyStores[0]
	d.SetId(aws.StringValue(keyStore.CustomKeyStoreId))
	d.Set("name", keyStore.CustomKeyStoreName)
	d.Set("cloudhsm_cluster_id", keyStore.CloudHsmClusterId)
	d.Set("connection_state", keyStore.ConnectionState)
	d.Set("creation_date", keyStore.CreationDate.Format(time.RFC3339))
	d.Set("trust_anchor_certificate", keyStore.TrustAnchorCertificate)

	return nil
}
