package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ConfigConformancePackCreateTimeout = 5 * time.Minute
	ConfigConformancePackDeleteTimeout = 5 * time.Minute

	ConfigConformancePackStatusNotFound = "NotFound"
	ConfigConformancePackStatusUnknown  = "Unknown"
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

func configDescribeConformancePackStatus(conn *configservice.ConfigService, name string) (*configservice.ConformancePackStatusDetail, error) {
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

func configDescribeOrganizationConfigRuleStatus(conn *configservice.ConfigService, name string) (*configservice.OrganizationConfigRuleStatus, error) {
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

func configDescribeOrganizationConformancePackStatus(conn *configservice.ConfigService, name string) (*configservice.OrganizationConformancePackStatus, error) {
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

func configGetOrganizationConfigRuleDetailedStatus(conn *configservice.ConfigService, ruleName, ruleStatus string) ([]*configservice.MemberAccountStatus, error) {
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

func configGetOrganizationConformancePackDetailedStatus(conn *configservice.ConfigService, name, status string) ([]*configservice.OrganizationConformancePackDetailedStatus, error) {
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

func configRefreshConformancePackStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := configDescribeConformancePackStatus(conn, name)

		if err != nil {
			return nil, ConfigConformancePackStatusUnknown, err
		}

		if status == nil {
			return nil, ConfigConformancePackStatusNotFound, nil
		}

		if errMsg := aws.StringValue(status.ConformancePackStatusReason); errMsg != "" {
			return status, aws.StringValue(status.ConformancePackState), fmt.Errorf(errMsg)
		}

		return status, aws.StringValue(status.ConformancePackState), nil
	}
}

func configRefreshOrganizationConfigRuleStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := configDescribeOrganizationConfigRuleStatus(conn, name)

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
			memberAccountStatuses, err := configGetOrganizationConfigRuleDetailedStatus(conn, name, aws.StringValue(status.OrganizationRuleStatus))

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

func configRefreshOrganizationConformancePackCreationStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := configDescribeOrganizationConformancePackStatus(conn, name)

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
			return status, s, configOrganizationConformancePackDetailedStatusError(conn, name, s)
		}

		return status, aws.StringValue(status.Status), nil
	}
}

func configRefreshOrganizationConformancePackStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := configDescribeOrganizationConformancePackStatus(conn, name)

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
			return status, s, configOrganizationConformancePackDetailedStatusError(conn, name, s)
		}

		return status, aws.StringValue(status.Status), nil
	}
}

func configOrganizationConformancePackDetailedStatusError(conn *configservice.ConfigService, name, status string) error {
	memberAccountStatuses, err := configGetOrganizationConformancePackDetailedStatus(conn, name, status)

	if err != nil {
		return fmt.Errorf("unable to get Config Organization Conformance Pack detailed status for showing member account errors: %w", err)
	}

	var errBuilder strings.Builder

	for _, mas := range memberAccountStatuses {
		errBuilder.WriteString(fmt.Sprintf("Account ID (%s): %s: %s\n", aws.StringValue(mas.AccountId), aws.StringValue(mas.ErrorCode), aws.StringValue(mas.ErrorMessage)))
	}

	return fmt.Errorf("Failed in %d account(s):\n\n%s", len(memberAccountStatuses), errBuilder.String())
}

func configWaitForConformancePackStateCreateComplete(conn *configservice.ConfigService, name string) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.ConformancePackStateCreateInProgress},
		Target:  []string{configservice.ConformancePackStateCreateComplete},
		Timeout: ConfigConformancePackCreateTimeout,
		Refresh: configRefreshConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
		return nil
	}

	return err

}

func configWaitForConformancePackStateDeleteComplete(conn *configservice.ConfigService, name string) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.ConformancePackStateDeleteInProgress},
		Target:  []string{},
		Timeout: ConfigConformancePackDeleteTimeout,
		Refresh: configRefreshConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
		return nil
	}

	return err
}

func configWaitForOrganizationConformancePackStatusCreateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.OrganizationResourceStatusCreateInProgress},
		Target:  []string{configservice.OrganizationResourceStatusCreateSuccessful},
		Timeout: timeout,
		Refresh: configRefreshOrganizationConformancePackCreationStatus(conn, name),
		// Include a delay to help avoid ResourceDoesNotExist errors
		Delay: 30 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err

}

func configWaitForOrganizationConformancePackStatusUpdateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.OrganizationResourceStatusUpdateInProgress},
		Target:  []string{configservice.OrganizationResourceStatusUpdateSuccessful},
		Timeout: timeout,
		Refresh: configRefreshOrganizationConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func configWaitForOrganizationConformancePackStatusDeleteSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := resource.StateChangeConf{
		Pending: []string{configservice.OrganizationResourceStatusDeleteInProgress},
		Target:  []string{configservice.OrganizationResourceStatusDeleteSuccessful},
		Timeout: timeout,
		Refresh: configRefreshOrganizationConformancePackStatus(conn, name),
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func configWaitForOrganizationRuleStatusCreateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusCreateInProgress},
		Target:  []string{configservice.OrganizationRuleStatusCreateSuccessful},
		Refresh: configRefreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func configWaitForOrganizationRuleStatusDeleteSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusDeleteInProgress},
		Target:  []string{configservice.OrganizationRuleStatusDeleteSuccessful},
		Refresh: configRefreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	if tfawserr.ErrMessageContains(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException, "") {
		return nil
	}

	return err
}

func configWaitForOrganizationRuleStatusUpdateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusUpdateInProgress},
		Target:  []string{configservice.OrganizationRuleStatusUpdateSuccessful},
		Refresh: configRefreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err
}
