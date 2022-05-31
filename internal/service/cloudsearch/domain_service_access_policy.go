package cloudsearch

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainServiceAccessPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainServiceAccessPolicyPut,
		Read:   resourceDomainServiceAccessPolicyRead,
		Update: resourceDomainServiceAccessPolicyPut,
		Delete: resourceDomainServiceAccessPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDomainServiceAccessPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	domainName := d.Get("domain_name").(string)
	input := &cloudsearch.UpdateServiceAccessPoliciesInput{
		DomainName: aws.String(domainName),
	}

	accessPolicy := d.Get("access_policy").(string)
	policy, err := structure.NormalizeJsonString(accessPolicy)

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", accessPolicy, err)
	}

	input.AccessPolicies = aws.String(policy)

	log.Printf("[DEBUG] Updating CloudSearch Domain access policies: %s", input)
	_, err = conn.UpdateServiceAccessPolicies(input)

	if err != nil {
		return fmt.Errorf("error updating CloudSearch Domain Service Access Policy (%s): %w", domainName, err)
	}

	d.SetId(domainName)

	_, err = waitAccessPolicyActive(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return fmt.Errorf("error waiting for CloudSearch Domain Service Access Policy (%s) to become active: %w", d.Id(), err)
	}

	return resourceDomainServiceAccessPolicyRead(d, meta)
}

func resourceDomainServiceAccessPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	accessPolicy, err := FindAccessPolicyByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudSearch Domain Service Access Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudSearch Domain Service Access Policy (%s): %w", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("access_policy").(string), accessPolicy)

	if err != nil {
		return err
	}

	d.Set("access_policy", policyToSet)
	d.Set("domain_name", d.Id())

	return nil
}

func resourceDomainServiceAccessPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	input := &cloudsearch.UpdateServiceAccessPoliciesInput{
		AccessPolicies: aws.String(""),
		DomainName:     aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting CloudSearch Domain Service Access Policy: %s", d.Id())
	_, err := conn.UpdateServiceAccessPolicies(input)

	if tfawserr.ErrCodeEquals(err, cloudsearch.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudSearch Domain Service Access Policy (%s): %w", d.Id(), err)
	}

	_, err = waitAccessPolicyActive(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for CloudSearch Domain Service Access Policy (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func FindAccessPolicyByName(conn *cloudsearch.CloudSearch, name string) (string, error) {
	output, err := findAccessPoliciesStatusByName(conn, name)

	if err != nil {
		return "", err
	}

	accessPolicy := aws.StringValue(output.Options)

	if accessPolicy == "" {
		return "", tfresource.NewEmptyResultError(name)
	}

	return accessPolicy, nil
}

func findAccessPoliciesStatusByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.AccessPoliciesStatus, error) {
	input := &cloudsearch.DescribeServiceAccessPoliciesInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeServiceAccessPolicies(input)

	if tfawserr.ErrCodeEquals(err, cloudsearch.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccessPolicies == nil || output.AccessPolicies.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccessPolicies, nil
}

func statusAccessPolicyState(conn *cloudsearch.CloudSearch, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccessPoliciesStatusByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.State), nil
	}
}

func waitAccessPolicyActive(conn *cloudsearch.CloudSearch, name string, timeout time.Duration) (*cloudsearch.AccessPoliciesStatus, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudsearch.OptionStateProcessing},
		Target:  []string{cloudsearch.OptionStateActive},
		Refresh: statusAccessPolicyState(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloudsearch.AccessPoliciesStatus); ok {
		return output, err
	}

	return nil, err
}
