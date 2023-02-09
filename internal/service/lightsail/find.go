package lightsail

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCertificateByName(ctx context.Context, conn *lightsail.Lightsail, name string) (*lightsail.Certificate, error) {
	in := &lightsail.GetCertificatesInput{
		CertificateName: aws.String(name),
	}

	out, err := conn.GetCertificatesWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Certificates) == 0 || out.Certificates[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Certificates[0].CertificateDetail, nil
}

func FindContainerServiceByName(ctx context.Context, conn *lightsail.Lightsail, serviceName string) (*lightsail.ContainerService, error) {
	input := &lightsail.GetContainerServicesInput{
		ServiceName: aws.String(serviceName),
	}

	output, err := conn.GetContainerServicesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ContainerServices) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ContainerServices); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ContainerServices[0], nil
}

func FindContainerServiceDeploymentByVersion(ctx context.Context, conn *lightsail.Lightsail, serviceName string, version int) (*lightsail.ContainerServiceDeployment, error) {
	input := &lightsail.GetContainerServiceDeploymentsInput{
		ServiceName: aws.String(serviceName),
	}

	output, err := conn.GetContainerServiceDeploymentsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Deployments) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	var result *lightsail.ContainerServiceDeployment

	for _, deployment := range output.Deployments {
		if deployment == nil {
			continue
		}

		if int(aws.Int64Value(deployment.Version)) == version {
			result = deployment
			break
		}
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return result, nil
}

func FindDiskById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.Disk, error) {
	in := &lightsail.GetDiskInput{
		DiskName: aws.String(id),
	}

	out, err := conn.GetDiskWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Disk == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Disk, nil
}

func FindDiskAttachmentById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.Disk, error) {
	id_parts := strings.SplitN(id, ",", -1)

	if len(id_parts) != 2 {
		return nil, errors.New("invalid Disk Attachment id")
	}

	dName := id_parts[0]
	iName := id_parts[1]

	in := &lightsail.GetDiskInput{
		DiskName: aws.String(dName),
	}

	out, err := conn.GetDiskWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	disk := out.Disk

	if disk == nil || !aws.BoolValue(disk.IsAttached) || aws.StringValue(disk.Name) != dName || aws.StringValue(disk.AttachedTo) != iName {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Disk, nil
}

func FindDomainEntryById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.DomainEntry, error) {
	id_parts := strings.SplitN(id, "_", -1)
	domainName := id_parts[1]
	name := expandDomainEntryName(id_parts[0], domainName)
	recordType := id_parts[2]
	recordTarget := id_parts[3]

	in := &lightsail.GetDomainInput{
		DomainName: aws.String(domainName),
	}

	if len(id_parts) != 4 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	out, err := conn.GetDomainWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry *lightsail.DomainEntry
	entryExists := false

	for _, n := range out.Domain.DomainEntries {
		if name == aws.StringValue(n.Name) && recordType == aws.StringValue(n.Type) && recordTarget == aws.StringValue(n.Target) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return entry, nil
}

func FindLoadBalancerByName(ctx context.Context, conn *lightsail.Lightsail, name string) (*lightsail.LoadBalancer, error) {
	in := &lightsail.GetLoadBalancerInput{LoadBalancerName: aws.String(name)}
	out, err := conn.GetLoadBalancerWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.LoadBalancer == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	lb := out.LoadBalancer

	return lb, nil
}

func FindLoadBalancerAttachmentById(ctx context.Context, conn *lightsail.Lightsail, id string) (*string, error) {
	id_parts := strings.SplitN(id, ",", -1)
	if len(id_parts) != 2 {
		return nil, errors.New("invalid load balancer attachment id")
	}

	lbName := id_parts[0]
	iName := id_parts[1]

	in := &lightsail.GetLoadBalancerInput{LoadBalancerName: aws.String(lbName)}
	out, err := conn.GetLoadBalancerWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry *string
	entryExists := false

	for _, n := range out.LoadBalancer.InstanceHealthSummary {
		if iName == aws.StringValue(n.InstanceName) {
			entry = n.InstanceName
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return entry, nil
}

func FindLoadBalancerCertificateById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.LoadBalancerTlsCertificate, error) {
	id_parts := strings.SplitN(id, ",", -1)
	if len(id_parts) != 2 {
		return nil, errors.New("invalid load balancer certificate id")
	}

	lbName := id_parts[0]
	cName := id_parts[1]

	in := &lightsail.GetLoadBalancerTlsCertificatesInput{LoadBalancerName: aws.String(lbName)}
	out, err := conn.GetLoadBalancerTlsCertificatesWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry *lightsail.LoadBalancerTlsCertificate
	entryExists := false

	for _, n := range out.TlsCertificates {
		if cName == aws.StringValue(n.Name) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return entry, nil
}

func FindLoadBalancerCertificateAttachmentById(ctx context.Context, conn *lightsail.Lightsail, id string) (*string, error) {
	id_parts := strings.SplitN(id, ",", -1)
	if len(id_parts) != 2 {
		return nil, errors.New("invalid load balancer certificate attachment id")
	}

	lbName := id_parts[0]
	cName := id_parts[1]

	in := &lightsail.GetLoadBalancerTlsCertificatesInput{LoadBalancerName: aws.String(lbName)}
	out, err := conn.GetLoadBalancerTlsCertificatesWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry *string
	entryExists := false

	for _, n := range out.TlsCertificates {
		if cName == aws.StringValue(n.Name) && aws.BoolValue(n.IsAttached) {
			entry = n.Name
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return entry, nil
}

func FindLoadBalancerStickinessPolicyById(ctx context.Context, conn *lightsail.Lightsail, id string) (map[string]*string, error) {
	in := &lightsail.GetLoadBalancerInput{LoadBalancerName: aws.String(id)}
	out, err := conn.GetLoadBalancerWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.LoadBalancer.ConfigurationOptions == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.LoadBalancer.ConfigurationOptions, nil
}

func FindLoadBalancerHTTPSRedirectionPolicyById(ctx context.Context, conn *lightsail.Lightsail, id string) (*bool, error) {
	in := &lightsail.GetLoadBalancerInput{LoadBalancerName: aws.String(id)}
	out, err := conn.GetLoadBalancerWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.LoadBalancer.HttpsRedirectionEnabled == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.LoadBalancer.HttpsRedirectionEnabled, nil
}

func FindBucketById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.Bucket, error) {
	in := &lightsail.GetBucketsInput{BucketName: aws.String(id)}
	out, err := conn.GetBucketsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Buckets) == 0 || out.Buckets[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Buckets[0], nil
}

func FindInstanceById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.Instance, error) {
	in := &lightsail.GetInstanceInput{InstanceName: aws.String(id)}
	out, err := conn.GetInstanceWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Instance == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Instance, nil
}

func FindBucketAccessKeyById(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.AccessKey, error) {
	parts, err := flex.ExpandResourceId(id, BucketAccessKeyIdPartsCount)

	if err != nil {
		return nil, err
	}

	in := &lightsail.GetBucketAccessKeysInput{BucketName: aws.String(parts[0])}
	out, err := conn.GetBucketAccessKeysWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry *lightsail.AccessKey
	entryExists := false

	for _, n := range out.AccessKeys {
		if parts[1] == aws.StringValue(n.AccessKeyId) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return entry, nil
}
