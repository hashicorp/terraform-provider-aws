package iam

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindGroupAttachedPolicy returns the AttachedPolicy corresponding to the specified group and policy ARN.
func FindGroupAttachedPolicy(conn *iam.IAM, groupName string, policyARN string) (*iam.AttachedPolicy, error) {
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	var result *iam.AttachedPolicy

	err := conn.ListAttachedGroupPoliciesPages(input, func(page *iam.ListAttachedGroupPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, attachedPolicy := range page.AttachedPolicies {
			if attachedPolicy == nil {
				continue
			}

			if aws.StringValue(attachedPolicy.PolicyArn) == policyARN {
				result = attachedPolicy
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindUserAttachedPolicy returns the AttachedPolicy corresponding to the specified user and policy ARN.
func FindUserAttachedPolicy(conn *iam.IAM, userName string, policyARN string) (*iam.AttachedPolicy, error) {
	input := &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	}

	var result *iam.AttachedPolicy

	err := conn.ListAttachedUserPoliciesPages(input, func(page *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, attachedPolicy := range page.AttachedPolicies {
			if attachedPolicy == nil {
				continue
			}

			if aws.StringValue(attachedPolicy.PolicyArn) == policyARN {
				result = attachedPolicy
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindPolicyByARN returns the Policy corresponding to the specified ARN.
func FindPolicyByARN(conn *iam.IAM, arn string) (*iam.Policy, error) {
	input := &iam.GetPolicyInput{
		PolicyArn: aws.String(arn),
	}

	output, err := conn.GetPolicy(input)
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}

// FindPolicyByName returns the Policy corresponding to the name and/or path-prefix.
func FindPolicyByName(conn *iam.IAM, name, pathPrefix string) (*iam.Policy, error) {
	input := &iam.ListPoliciesInput{}
	if pathPrefix != "" {
		input.PathPrefix = aws.String(pathPrefix)
	}

	all, err := FindPolicies(conn, input)
	if err != nil {
		return nil, err
	}

	var results []*iam.Policy
	for _, p := range all {
		if name != "" && name != aws.StringValue(p.PolicyName) {
			continue
		}

		results = append(results, p)
	}

	if len(results) == 0 || results[0] == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}
	if l := len(results); l > 1 {
		return nil, tfresource.NewTooManyResultsError(1, nil)
	}

	return results[0], nil
}

// FindPolicies returns the Policies matching the parameters in the iam.ListPoliciesInput.
func FindPolicies(conn *iam.IAM, input *iam.ListPoliciesInput) ([]*iam.Policy, error) {
	var results []*iam.Policy
	err := conn.ListPoliciesPages(input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, p := range page.Policies {
			if p == nil {
				continue
			}

			results = append(results, p)
		}

		return !lastPage
	})

	return results, err
}

func FindUsers(conn *iam.IAM, nameRegex, pathPrefix string) ([]*iam.User, error) {
	input := &iam.ListUsersInput{}

	if pathPrefix != "" {
		input.PathPrefix = aws.String(pathPrefix)
	}

	var results []*iam.User

	err := conn.ListUsersPages(input, func(page *iam.ListUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, user := range page.Users {
			if user == nil {
				continue
			}

			if nameRegex != "" && !regexp.MustCompile(nameRegex).MatchString(aws.StringValue(user.UserName)) {
				continue
			}

			results = append(results, user)
		}

		return !lastPage
	})

	return results, err
}

func FindRoleByName(conn *iam.IAM, name string) (*iam.Role, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	output, err := conn.GetRole(input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Role == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Role, nil
}

func FindVirtualMFADevice(conn *iam.IAM, serialNum string) (*iam.VirtualMFADevice, error) {
	input := &iam.ListVirtualMFADevicesInput{}

	output, err := conn.ListVirtualMFADevices(input)

	if err != nil {
		return nil, err
	}

	if len(output.VirtualMFADevices) == 0 || output.VirtualMFADevices[0] == nil {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var device *iam.VirtualMFADevice

	for _, dvs := range output.VirtualMFADevices {
		if aws.StringValue(dvs.SerialNumber) == serialNum {
			device = dvs
			break
		}
	}

	if device == nil {
		return nil, tfresource.NewEmptyResultError(device)
	}

	return device, nil
}

func FindServiceSpecificCredential(conn *iam.IAM, serviceName, userName, credID string) (*iam.ServiceSpecificCredentialMetadata, error) {
	input := &iam.ListServiceSpecificCredentialsInput{
		ServiceName: aws.String(serviceName),
		UserName:    aws.String(userName),
	}

	output, err := conn.ListServiceSpecificCredentials(input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.ServiceSpecificCredentials) == 0 || output.ServiceSpecificCredentials[0] == nil {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var cred *iam.ServiceSpecificCredentialMetadata

	for _, crd := range output.ServiceSpecificCredentials {
		if aws.StringValue(crd.ServiceName) == serviceName &&
			aws.StringValue(crd.UserName) == userName &&
			aws.StringValue(crd.ServiceSpecificCredentialId) == credID {
			cred = crd
			break
		}
	}

	if cred == nil {
		return nil, tfresource.NewEmptyResultError(cred)
	}

	return cred, nil
}

func FindSigningCertificate(conn *iam.IAM, userName, certId string) (*iam.SigningCertificate, error) {
	input := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}

	output, err := conn.ListSigningCertificates(input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.Certificates) == 0 || output.Certificates[0] == nil {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var cert *iam.SigningCertificate

	for _, crt := range output.Certificates {
		if aws.StringValue(crt.UserName) == userName &&
			aws.StringValue(crt.CertificateId) == certId {
			cert = crt
			break
		}
	}

	if cert == nil {
		return nil, tfresource.NewEmptyResultError(cert)
	}

	return cert, nil
}

func FindSAMLProviderByARN(ctx context.Context, conn *iam.IAM, arn string) (*iam.GetSAMLProviderOutput, error) {
	input := &iam.GetSAMLProviderInput{
		SAMLProviderArn: aws.String(arn),
	}

	output, err := conn.GetSAMLProviderWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindAccessKey(ctx context.Context, conn *iam.IAM, username, id string) (*iam.AccessKeyMetadata, error) {
	accessKeys, err := FindAccessKeys(ctx, conn, username)
	if err != nil {
		return nil, err
	}

	for _, accessKey := range accessKeys {
		if aws.StringValue(accessKey.AccessKeyId) == id {
			return accessKey, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindAccessKeys(ctx context.Context, conn *iam.IAM, username string) ([]*iam.AccessKeyMetadata, error) {
	var accessKeys []*iam.AccessKeyMetadata
	input := &iam.ListAccessKeysInput{
		UserName: aws.String(username),
	}
	err := conn.ListAccessKeysPages(input, func(page *iam.ListAccessKeysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		accessKeys = append(accessKeys, page.AccessKeyMetadata...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	return accessKeys, err
}
