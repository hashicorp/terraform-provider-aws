package configservice

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	conformancePackCreateTimeout = 5 * time.Minute
	conformancePackDeleteTimeout = 5 * time.Minute

	conformancePackStatusNotFound = "NotFound"
	conformancePackStatusUnknown  = "Unknown"
)

func DescribeConformancePack(conn *configservice.ConfigService, name string) (*configservice.ConformancePackDetail, error) {
	input := &configservice.DescribeConformancePacksInput{
		ConformancePackNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeConformancePacks(input)

		if err != nil {
			return nil, err
		}

		for _, pack := range output.ConformancePackDetails {
			if pack == nil {
				continue
			}

			if aws.StringValue(pack.ConformancePackName) == name {
				return pack, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func describeConformancePackStatus(conn *configservice.ConfigService, name string) (*configservice.ConformancePackStatusDetail, error) {
	input := &configservice.DescribeConformancePackStatusInput{
		ConformancePackNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeConformancePackStatus(input)

		if err != nil {
			return nil, err
		}

		for _, status := range output.ConformancePackStatusDetails {
			if aws.StringValue(status.ConformancePackName) == name {
				return status, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func DescribeOrganizationConfigRule(conn *configservice.ConfigService, name string) (*configservice.OrganizationConfigRule, error) {
	input := &configservice.DescribeOrganizationConfigRulesInput{
		OrganizationConfigRuleNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeOrganizationConfigRules(input)

		if err != nil {
			return nil, err
		}

		for _, rule := range output.OrganizationConfigRules {
			if aws.StringValue(rule.OrganizationConfigRuleName) == name {
				return rule, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func describeOrganizationConfigRuleStatus(conn *configservice.ConfigService, name string) (*configservice.OrganizationConfigRuleStatus, error) {
	input := &configservice.DescribeOrganizationConfigRuleStatusesInput{
		OrganizationConfigRuleNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeOrganizationConfigRuleStatuses(input)

		if err != nil {
			return nil, err
		}

		for _, status := range output.OrganizationConfigRuleStatuses {
			if aws.StringValue(status.OrganizationConfigRuleName) == name {
				return status, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func DescribeOrganizationConformancePack(conn *configservice.ConfigService, name string) (*configservice.OrganizationConformancePack, error) {
	input := &configservice.DescribeOrganizationConformancePacksInput{
		OrganizationConformancePackNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeOrganizationConformancePacks(input)

		if err != nil {
			return nil, err
		}

		for _, pack := range output.OrganizationConformancePacks {
			if aws.StringValue(pack.OrganizationConformancePackName) == name {
				return pack, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func describeOrganizationConformancePackStatus(conn *configservice.ConfigService, name string) (*configservice.OrganizationConformancePackStatus, error) {
	input := &configservice.DescribeOrganizationConformancePackStatusesInput{
		OrganizationConformancePackNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeOrganizationConformancePackStatuses(input)

		if err != nil {
			return nil, err
		}

		for _, status := range output.OrganizationConformancePackStatuses {
			if aws.StringValue(status.OrganizationConformancePackName) == name {
				return status, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func getOrganizationConfigRuleDetailedStatus(conn *configservice.ConfigService, ruleName, ruleStatus string) ([]*configservice.MemberAccountStatus, error) {
	input := &configservice.GetOrganizationConfigRuleDetailedStatusInput{
		Filters: &configservice.StatusDetailFilters{
			MemberAccountRuleStatus: aws.String(ruleStatus),
		},
		OrganizationConfigRuleName: aws.String(ruleName),
	}
	var statuses []*configservice.MemberAccountStatus

	for {
		output, err := conn.GetOrganizationConfigRuleDetailedStatus(input)

		if err != nil {
			return nil, err
		}

		statuses = append(statuses, output.OrganizationConfigRuleDetailedStatus...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return statuses, nil
}

func getOrganizationConformancePackDetailedStatus(conn *configservice.ConfigService, name, status string) ([]*configservice.OrganizationConformancePackDetailedStatus, error) {
	input := &configservice.GetOrganizationConformancePackDetailedStatusInput{
		Filters: &configservice.OrganizationResourceDetailedStatusFilters{
			Status: aws.String(status),
		},
		OrganizationConformancePackName: aws.String(name),
	}

	var statuses []*configservice.OrganizationConformancePackDetailedStatus

	for {
		output, err := conn.GetOrganizationConformancePackDetailedStatus(input)

		if err != nil {
			return nil, err
		}

		statuses = append(statuses, output.OrganizationConformancePackDetailedStatuses...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return statuses, nil
}

func refreshConformancePackStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := describeConformancePackStatus(conn, name)

		if err != nil {
			return nil, conformancePackStatusUnknown, err
		}

		if status == nil {
			return nil, conformancePackStatusNotFound, nil
		}

		if errMsg := aws.StringValue(status.ConformancePackStatusReason); errMsg != "" {
			return status, aws.StringValue(status.ConformancePackState), fmt.Errorf(errMsg)
		}

		return status, aws.StringValue(status.ConformancePackState), nil
	}
}

func refreshOrganizationConfigRuleStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := describeOrganizationConfigRuleStatus(conn, name)

		if err != nil {
			return nil, "", err
		}

		if status == nil {
			return nil, "", fmt.Errorf("status not found")
		}

		if status.ErrorCode != nil {
			return status, aws.StringValue(status.OrganizationRuleStatus), fmt.Errorf("%s: %s", aws.StringValue(status.ErrorCode), aws.StringValue(status.ErrorMessage))
		}

		switch aws.StringValue(status.OrganizationRuleStatus) {
		case configservice.OrganizationRuleStatusCreateFailed, configservice.OrganizationRuleStatusDeleteFailed, configservice.OrganizationRuleStatusUpdateFailed:
			// Display detailed errors for failed member accounts
			memberAccountStatuses, err := getOrganizationConfigRuleDetailedStatus(conn, name, aws.StringValue(status.OrganizationRuleStatus))

			if err != nil {
				return status, aws.StringValue(status.OrganizationRuleStatus), fmt.Errorf("unable to get Organization Config Rule detailed status for showing member account errors: %w", err)
			}

			var errBuilder strings.Builder

			for _, mas := range memberAccountStatuses {
				errBuilder.WriteString(fmt.Sprintf("Account ID (%s): %s: %s\n", aws.StringValue(mas.AccountId), aws.StringValue(mas.ErrorCode), aws.StringValue(mas.ErrorMessage)))
			}

			return status, aws.StringValue(status.OrganizationRuleStatus), fmt.Errorf("Failed in %d account(s):\n\n%s", len(memberAccountStatuses), errBuilder.String())
		}

		return status, aws.StringValue(status.OrganizationRuleStatus), nil
	}
}

func refreshOrganizationConformancePackCreationStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := describeOrganizationConformancePackStatus(conn, name)

		// Transient ResourceDoesNotExist error after creation caught here
		// in cases where the StateChangeConf's delay time is not sufficient
		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if status == nil {
			return nil, "", nil
		}

		if status.ErrorCode != nil {
			return status, aws.StringValue(status.Status), fmt.Errorf("%s: %s", aws.StringValue(status.ErrorCode), aws.StringValue(status.ErrorMessage))
		}

		switch s := aws.StringValue(status.Status); s {
		case configservice.OrganizationResourceStatusCreateFailed, configservice.OrganizationResourceStatusDeleteFailed, configservice.OrganizationResourceStatusUpdateFailed:
			return status, s, organizationConformancePackDetailedStatusError(conn, name, s)
		}

		return status, aws.StringValue(status.Status), nil
	}
}

func refreshOrganizationConformancePackStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := describeOrganizationConformancePackStatus(conn, name)

		if err != nil {
			return nil, "", err
		}

		if status == nil {
			return nil, "", nil
		}

		if status.ErrorCode != nil {
			return status, aws.StringValue(status.Status), fmt.Errorf("%s: %s", aws.StringValue(status.ErrorCode), aws.StringValue(status.ErrorMessage))
		}

		switch s := aws.StringValue(status.Status); s {
		case configservice.OrganizationResourceStatusCreateFailed, configservice.OrganizationResourceStatusDeleteFailed, configservice.OrganizationResourceStatusUpdateFailed:
			return status, s, organizationConformancePackDetailedStatusError(conn, name, s)
		}

		return status, aws.StringValue(status.Status), nil
	}
}

func organizationConformancePackDetailedStatusError(conn *configservice.ConfigService, name, status string) error {
	memberAccountStatuses, err := getOrganizationConformancePackDetailedStatus(conn, name, status)

	if err != nil {
		return fmt.Errorf("unable to get Config Organization Conformance Pack detailed status for showing member account errors: %w", err)
	}

	var errBuilder strings.Builder

	for _, mas := range memberAccountStatuses {
		errBuilder.WriteString(fmt.Sprintf("Account ID (%s): %s: %s\n", aws.StringValue(mas.AccountId), aws.StringValue(mas.ErrorCode), aws.StringValue(mas.ErrorMessage)))
	}

	return fmt.Errorf("Failed in %d account(s):\n\n%s", len(memberAccountStatuses), errBuilder.String())
}

func waitForConformancePackStateCreateComplete(conn *configservice.ConfigService, name string) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.ConformancePackStateCreateInProgress},
		Target:  []string{configservice.ConformancePackStateCreateComplete},
		Timeout: conformancePackCreateTimeout,
		Refresh: refreshConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
		return nil
	}

	return err
}

func waitForConformancePackStateDeleteComplete(conn *configservice.ConfigService, name string) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.ConformancePackStateDeleteInProgress},
		Target:  []string{},
		Timeout: conformancePackDeleteTimeout,
		Refresh: refreshConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
		return nil
	}

	return err
}

func waitForOrganizationConformancePackStatusCreateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.OrganizationResourceStatusCreateInProgress},
		Target:  []string{configservice.OrganizationResourceStatusCreateSuccessful},
		Timeout: timeout,
		Refresh: refreshOrganizationConformancePackCreationStatus(conn, name),
		// Include a delay to help avoid ResourceDoesNotExist errors
		Delay: 30 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func waitForOrganizationConformancePackStatusUpdateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.OrganizationResourceStatusUpdateInProgress},
		Target:  []string{configservice.OrganizationResourceStatusUpdateSuccessful},
		Timeout: timeout,
		Refresh: refreshOrganizationConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func waitForOrganizationConformancePackStatusDeleteSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.OrganizationResourceStatusDeleteInProgress},
		Target:  []string{configservice.OrganizationResourceStatusDeleteSuccessful},
		Timeout: timeout,
		Refresh: refreshOrganizationConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func waitForOrganizationRuleStatusCreateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusCreateInProgress},
		Target:  []string{configservice.OrganizationRuleStatusCreateSuccessful},
		Refresh: refreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func waitForOrganizationRuleStatusDeleteSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusDeleteInProgress},
		Target:  []string{configservice.OrganizationRuleStatusDeleteSuccessful},
		Refresh: refreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException) {
		return nil
	}

	return err
}

func waitForOrganizationRuleStatusUpdateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusUpdateInProgress},
		Target:  []string{configservice.OrganizationRuleStatusUpdateSuccessful},
		Refresh: refreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err
}
