package aws

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
)

func resourceAwsDirectoryServiceConditionalForwarder() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDirectoryServiceConditionalForwarderCreate,
		Read:   resourceAwsDirectoryServiceConditionalForwarderRead,
		Update: resourceAwsDirectoryServiceConditionalForwarderUpdate,
		Delete: resourceAwsDirectoryServiceConditionalForwarderDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"directory_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dns_ips": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					//ValidateFunc: validation.SingleIP(),
				},
			},

			"domain_name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^([a-zA-Z0-9]+[\\.-])+([a-zA-Z0-9])+[.]?$"), "'domain_name' is incorrect"),
			},
		},
	}
}

func resourceAwsDirectoryServiceConditionalForwarderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dsconn

	var dnsIps []*string
	for _, ip := range d.Get("dns_ips").(*schema.Set).List() {
		dnsIps = append(dnsIps, aws.String(ip.(string)))
	}

	directoryId := d.Get("directory_id").(string)
	domainName := d.Get("domain_name").(string)

	_, err := conn.CreateConditionalForwarder(&directoryservice.CreateConditionalForwarderInput{
		DirectoryId:      aws.String(directoryId),
		DnsIpAddrs:       dnsIps,
		RemoteDomainName: aws.String(domainName),
	})

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s", directoryId, domainName))

	return nil
}

func resourceAwsDirectoryServiceConditionalForwarderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dsconn

	parts := strings.SplitN(d.Id(), ":", 2)

	if len(parts) != 2 {
		return fmt.Errorf("Incorrect id %q, expecting DIRECTORY_ID:DOMAIN_NAME", d.Id())
	}

	directoryId, domainName := parts[0], parts[1]

	res, err := conn.DescribeConditionalForwarders(&directoryservice.DescribeConditionalForwardersInput{
		DirectoryId:       aws.String(directoryId),
		RemoteDomainNames: []*string{aws.String(domainName)},
	})

	if err != nil {
		if isAWSErr(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	if len(res.ConditionalForwarders) == 0 {
		d.SetId("")
		return nil
	}

	cfd := res.ConditionalForwarders[0]

	d.Set("dns_ips", schema.NewSet(schema.HashString, flattenStringList(cfd.DnsIpAddrs)))
	d.Set("directory_id", directoryId)
	d.Set("domain_name", *cfd.RemoteDomainName)

	return nil
}

func resourceAwsDirectoryServiceConditionalForwarderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dsconn

	var dnsIps []*string
	for _, ip := range d.Get("dns_ips").(*schema.Set).List() {
		dnsIps = append(dnsIps, aws.String(ip.(string)))
	}

	_, err := conn.UpdateConditionalForwarder(&directoryservice.UpdateConditionalForwarderInput{
		DirectoryId:      aws.String(d.Get("directory_id").(string)),
		DnsIpAddrs:       dnsIps,
		RemoteDomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err != nil {
		return err
	}

	return resourceAwsDirectoryServiceConditionalForwarderRead(d, meta)
}

func resourceAwsDirectoryServiceConditionalForwarderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dsconn

	_, err := conn.DeleteConditionalForwarder(&directoryservice.DeleteConditionalForwarderInput{
		DirectoryId:      aws.String(d.Get("directory_id").(string)),
		RemoteDomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err != nil && !isAWSErr(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
		return err
	}

	d.SetId("")
	return nil
}
