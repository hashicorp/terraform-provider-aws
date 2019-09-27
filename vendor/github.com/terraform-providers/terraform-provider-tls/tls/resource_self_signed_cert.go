package tls

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/url"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceSelfSignedCert() *schema.Resource {
	s := resourceCertificateCommonSchema()

	s["subject"] = &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem:     nameSchema,
		ForceNew: true,
	}

	s["dns_names"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "List of DNS names to use as subjects of the certificate",
		ForceNew:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	s["ip_addresses"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "List of IP addresses to use as subjects of the certificate",
		ForceNew:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	s["uris"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "List of URIs to use as subjects of the certificate",
		ForceNew:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	s["key_algorithm"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the algorithm to use to generate the certificate's private key",
		ForceNew:    true,
	}

	s["private_key_pem"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "PEM-encoded private key that the certificate will belong to",
		ForceNew:    true,
		Sensitive:   true,
		StateFunc: func(v interface{}) string {
			return hashForState(v.(string))
		},
	}

	return &schema.Resource{
		Create:        CreateSelfSignedCert,
		Delete:        DeleteCertificate,
		Read:          ReadCertificate,
		Update:        UpdateCertificate,
		CustomizeDiff: CustomizeCertificateDiff,
		Schema:        s,
	}
}

func CreateSelfSignedCert(d *schema.ResourceData, meta interface{}) error {
	key, err := parsePrivateKey(d, "private_key_pem", "key_algorithm")
	if err != nil {
		return err
	}

	subjectConfs := d.Get("subject").([]interface{})
	if len(subjectConfs) != 1 {
		return fmt.Errorf("must have exactly one 'subject' block")
	}
	subjectConf, ok := subjectConfs[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("subject block cannot be empty")
	}
	subject, err := nameFromResourceData(subjectConf)
	if err != nil {
		return fmt.Errorf("invalid subject block: %s", err)
	}

	cert := x509.Certificate{
		Subject:               *subject,
		BasicConstraintsValid: true,
	}

	dnsNamesI := d.Get("dns_names").([]interface{})
	for _, nameI := range dnsNamesI {
		cert.DNSNames = append(cert.DNSNames, nameI.(string))
	}
	ipAddressesI := d.Get("ip_addresses").([]interface{})
	for _, ipStrI := range ipAddressesI {
		ip := net.ParseIP(ipStrI.(string))
		if ip == nil {
			return fmt.Errorf("invalid IP address %#v", ipStrI.(string))
		}
		cert.IPAddresses = append(cert.IPAddresses, ip)
	}
	urisI := d.Get("uris").([]interface{})
	for _, uriStrI := range urisI {
		uri, err := url.Parse(uriStrI.(string))
		if err != nil {
			return fmt.Errorf("invalid URI %#v", uriStrI.(string))
		}
		cert.URIs = append(cert.URIs, uri)
	}

	return createCertificate(d, &cert, &cert, publicKey(key), key)
}
