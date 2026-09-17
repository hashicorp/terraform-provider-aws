// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func containerDefinitionsAreEquivalent(def1, def2 string, isAWSVPC bool) (bool, error) {
	var obj1 containerDefinitions
	err := tfjson.DecodeFromString(def1, &obj1)
	if err != nil {
		return false, err
	}
	obj1.reduce(isAWSVPC)
	b1, err := tfjson.EncodeToBytes(obj1)
	if err != nil {
		return false, err
	}

	var obj2 containerDefinitions
	err = tfjson.DecodeFromString(def2, &obj2)
	if err != nil {
		return false, err
	}
	obj2.reduce(isAWSVPC)
	b2, err := tfjson.EncodeToBytes(obj2)
	if err != nil {
		return false, err
	}

	return tfjson.EqualBytes(b1, b2), nil
}

type containerDefinitions []awstypes.ContainerDefinition

func (cd containerDefinitions) reduce(isAWSVPC bool) {
	// Deal with fields which may be re-ordered in the API.
	cd.orderContainers()
	cd.orderEnvironmentVariables()
	cd.orderSecrets()

	// Compact any sparse lists.
	cd.compactArrays()

	// Deal with special fields which have defaults.
	// See https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#container_definitions.
	for i, def := range cd {
		if def.Essential == nil {
			cd[i].Essential = aws.Bool(true)
		}

		if hc := def.HealthCheck; hc != nil {
			if hc.Interval == nil {
				hc.Interval = aws.Int32(30)
			}
			if hc.Retries == nil {
				hc.Retries = aws.Int32(3)
			}
			if hc.Timeout == nil {
				hc.Timeout = aws.Int32(5)
			}
		}

		for j, pm := range def.PortMappings {
			if pm.Protocol == awstypes.TransportProtocolTcp {
				cd[i].PortMappings[j].Protocol = ""
			}
			if aws.ToInt32(pm.HostPort) == 0 {
				cd[i].PortMappings[j].HostPort = nil
			}
			if isAWSVPC && cd[i].PortMappings[j].HostPort == nil {
				cd[i].PortMappings[j].HostPort = cd[i].PortMappings[j].ContainerPort
			}
		}

		// Set all empty slices to nil.
		if len(def.Command) == 0 {
			cd[i].Command = nil
		}
		if len(def.CredentialSpecs) == 0 {
			cd[i].CredentialSpecs = nil
		}
		if len(def.DependsOn) == 0 {
			cd[i].DependsOn = nil
		}
		if len(def.DnsSearchDomains) == 0 {
			cd[i].DnsSearchDomains = nil
		}
		if len(def.DnsServers) == 0 {
			cd[i].DnsServers = nil
		}
		if len(def.DockerSecurityOptions) == 0 {
			cd[i].DockerSecurityOptions = nil
		}
		if len(def.EntryPoint) == 0 {
			cd[i].EntryPoint = nil
		}
		if len(def.Environment) == 0 {
			cd[i].Environment = nil
		}
		if len(def.EnvironmentFiles) == 0 {
			cd[i].EnvironmentFiles = nil
		}
		if len(def.ExtraHosts) == 0 {
			cd[i].ExtraHosts = nil
		}
		if len(def.Links) == 0 {
			cd[i].Links = nil
		}
		if len(def.MountPoints) == 0 {
			cd[i].MountPoints = nil
		}
		if len(def.PortMappings) == 0 {
			cd[i].PortMappings = nil
		}
		if len(def.ResourceRequirements) == 0 {
			cd[i].ResourceRequirements = nil
		}
		if len(def.Secrets) == 0 {
			cd[i].Secrets = nil
		}
		if len(def.SystemControls) == 0 {
			cd[i].SystemControls = nil
		}
		if len(def.Ulimits) == 0 {
			cd[i].Ulimits = nil
		}
		if len(def.VolumesFrom) == 0 {
			cd[i].VolumesFrom = nil
		}
	}
}

func (cd containerDefinitions) orderEnvironmentVariables() {
	for i, def := range cd {
		slices.SortFunc(def.Environment, func(a, b awstypes.KeyValuePair) int {
			return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
		})
		cd[i].Environment = def.Environment
	}
}

func (cd containerDefinitions) orderSecrets() {
	for i, def := range cd {
		slices.SortFunc(def.Secrets, func(a, b awstypes.Secret) int {
			return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
		})
		cd[i].Secrets = def.Secrets
	}
}

func (cd containerDefinitions) orderContainers() {
	slices.SortFunc(cd, func(a, b awstypes.ContainerDefinition) int {
		return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
	})
}

// compactArrays removes any zero values from the object arrays in the container definitions.
func (cd containerDefinitions) compactArrays() {
	for i, def := range cd {
		cd[i].DependsOn = compactArray(def.DependsOn)
		cd[i].Environment = compactArray(def.Environment)
		cd[i].EnvironmentFiles = compactArray(def.EnvironmentFiles)
		cd[i].ExtraHosts = compactArray(def.ExtraHosts)
		cd[i].MountPoints = compactArray(def.MountPoints)
		cd[i].PortMappings = compactArray(def.PortMappings)
		cd[i].ResourceRequirements = compactArray(def.ResourceRequirements)
		cd[i].Secrets = compactArray(def.Secrets)
		cd[i].SystemControls = compactArray(def.SystemControls)
		cd[i].Ulimits = compactArray(def.Ulimits)
		cd[i].VolumesFrom = compactArray(def.VolumesFrom)
	}
}

func compactArray[S ~[]E, E any](s S) S {
	if len(s) == 0 {
		return s
	}

	return tfslices.Filter(s, func(e E) bool {
		return !inttypes.IsZero(&e)
	})
}

func serializeContainerDefinitions(v []awstypes.ContainerDefinition, value smithyjson.Value) error {
	a := value.Array()
	defer a.Close()

	for _, container := range v {
		serializeContainerDefinition(&container, a.Value())
	}

	return nil
}

// Preserve the SDK wire representation used in state, including empty collections,
// omitted zero CPU and enums, and required zero-valued tmpfs sizes and ulimits.
func serializeContainerDefinition(v *awstypes.ContainerDefinition, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Command != nil {
		serializeStringList(v.Command, o.Key("command"))
	}
	if v.Cpu != 0 {
		o.Key("cpu").Integer(v.Cpu)
	}
	if v.CredentialSpecs != nil {
		serializeStringList(v.CredentialSpecs, o.Key("credentialSpecs"))
	}
	if v.DependsOn != nil {
		a := o.Key("dependsOn").Array()
		for _, dependency := range v.DependsOn {
			d := a.Value().Object()
			if dependency.Condition != "" {
				d.Key(names.AttrCondition).String(string(dependency.Condition))
			}
			if dependency.ContainerName != nil {
				d.Key("containerName").String(*dependency.ContainerName)
			}
			d.Close()
		}
		a.Close()
	}
	if v.DisableNetworking != nil {
		o.Key("disableNetworking").Boolean(*v.DisableNetworking)
	}
	if v.DnsSearchDomains != nil {
		serializeStringList(v.DnsSearchDomains, o.Key("dnsSearchDomains"))
	}
	if v.DnsServers != nil {
		serializeStringList(v.DnsServers, o.Key("dnsServers"))
	}
	if v.DockerLabels != nil {
		serializeStringMap(v.DockerLabels, o.Key("dockerLabels"))
	}
	if v.DockerSecurityOptions != nil {
		serializeStringList(v.DockerSecurityOptions, o.Key("dockerSecurityOptions"))
	}
	if v.EntryPoint != nil {
		serializeStringList(v.EntryPoint, o.Key("entryPoint"))
	}
	if v.Environment != nil {
		serializeEnvironment(v.Environment, o.Key(names.AttrEnvironment))
	}
	if v.EnvironmentFiles != nil {
		a := o.Key("environmentFiles").Array()
		for _, file := range v.EnvironmentFiles {
			f := a.Value().Object()
			if file.Type != "" {
				f.Key(names.AttrType).String(string(file.Type))
			}
			if file.Value != nil {
				f.Key(names.AttrValue).String(*file.Value)
			}
			f.Close()
		}
		a.Close()
	}
	if v.Essential != nil {
		o.Key("essential").Boolean(*v.Essential)
	}
	if v.ExtraHosts != nil {
		a := o.Key("extraHosts").Array()
		for _, host := range v.ExtraHosts {
			h := a.Value().Object()
			if host.Hostname != nil {
				h.Key("hostname").String(*host.Hostname)
			}
			if host.IpAddress != nil {
				h.Key("ipAddress").String(*host.IpAddress)
			}
			h.Close()
		}
		a.Close()
	}
	if v.FirelensConfiguration != nil {
		f := o.Key("firelensConfiguration").Object()
		if v.FirelensConfiguration.Options != nil {
			serializeStringMap(v.FirelensConfiguration.Options, f.Key("options"))
		}
		if v.FirelensConfiguration.Type != "" {
			f.Key(names.AttrType).String(string(v.FirelensConfiguration.Type))
		}
		f.Close()
	}
	if v.HealthCheck != nil {
		serializeHealthCheck(v.HealthCheck, o.Key("healthCheck"))
	}
	if v.Hostname != nil {
		o.Key("hostname").String(*v.Hostname)
	}
	if v.Image != nil {
		o.Key("image").String(*v.Image)
	}
	if v.Interactive != nil {
		o.Key("interactive").Boolean(*v.Interactive)
	}
	if v.Links != nil {
		serializeStringList(v.Links, o.Key("links"))
	}
	if v.LinuxParameters != nil {
		serializeLinuxParameters(v.LinuxParameters, o.Key("linuxParameters"))
	}
	if v.LogConfiguration != nil {
		serializeLogConfiguration(v.LogConfiguration, o.Key("logConfiguration"))
	}
	if v.Memory != nil {
		o.Key("memory").Integer(*v.Memory)
	}
	if v.MemoryReservation != nil {
		o.Key("memoryReservation").Integer(*v.MemoryReservation)
	}
	if v.MountPoints != nil {
		serializeMountPoints(v.MountPoints, o.Key("mountPoints"))
	}
	if v.Name != nil {
		o.Key(names.AttrName).String(*v.Name)
	}
	if v.PortMappings != nil {
		serializePortMappings(v.PortMappings, o.Key("portMappings"))
	}
	if v.Privileged != nil {
		o.Key("privileged").Boolean(*v.Privileged)
	}
	if v.PseudoTerminal != nil {
		o.Key("pseudoTerminal").Boolean(*v.PseudoTerminal)
	}
	if v.ReadonlyRootFilesystem != nil {
		o.Key("readonlyRootFilesystem").Boolean(*v.ReadonlyRootFilesystem)
	}
	if v.RepositoryCredentials != nil {
		r := o.Key("repositoryCredentials").Object()
		if v.RepositoryCredentials.CredentialsParameter != nil {
			r.Key("credentialsParameter").String(*v.RepositoryCredentials.CredentialsParameter)
		}
		r.Close()
	}
	if v.ResourceRequirements != nil {
		a := o.Key("resourceRequirements").Array()
		for _, requirement := range v.ResourceRequirements {
			r := a.Value().Object()
			if requirement.Type != "" {
				r.Key(names.AttrType).String(string(requirement.Type))
			}
			if requirement.Value != nil {
				r.Key(names.AttrValue).String(*requirement.Value)
			}
			r.Close()
		}
		a.Close()
	}
	if v.RestartPolicy != nil {
		serializeContainerRestartPolicy(v.RestartPolicy, o.Key("restartPolicy"))
	}
	if v.Secrets != nil {
		serializeSecrets(v.Secrets, o.Key("secrets"))
	}
	if v.StartTimeout != nil {
		o.Key("startTimeout").Integer(*v.StartTimeout)
	}
	if v.StopTimeout != nil {
		o.Key("stopTimeout").Integer(*v.StopTimeout)
	}
	if v.SystemControls != nil {
		a := o.Key("systemControls").Array()
		for _, control := range v.SystemControls {
			c := a.Value().Object()
			if control.Namespace != nil {
				c.Key(names.AttrNamespace).String(*control.Namespace)
			}
			if control.Value != nil {
				c.Key(names.AttrValue).String(*control.Value)
			}
			c.Close()
		}
		a.Close()
	}
	if v.Ulimits != nil {
		a := o.Key("ulimits").Array()
		for _, limit := range v.Ulimits {
			l := a.Value().Object()
			l.Key("hardLimit").Integer(limit.HardLimit)
			if limit.Name != "" {
				l.Key(names.AttrName).String(string(limit.Name))
			}
			l.Key("softLimit").Integer(limit.SoftLimit)
			l.Close()
		}
		a.Close()
	}
	if v.User != nil {
		o.Key("user").String(*v.User)
	}
	if v.VersionConsistency != "" {
		o.Key("versionConsistency").String(string(v.VersionConsistency))
	}
	if v.VolumesFrom != nil {
		a := o.Key("volumesFrom").Array()
		for _, volume := range v.VolumesFrom {
			f := a.Value().Object()
			if volume.ReadOnly != nil {
				f.Key("readOnly").Boolean(*volume.ReadOnly)
			}
			if volume.SourceContainer != nil {
				f.Key("sourceContainer").String(*volume.SourceContainer)
			}
			f.Close()
		}
		a.Close()
	}
	if v.WorkingDirectory != nil {
		o.Key("workingDirectory").String(*v.WorkingDirectory)
	}
}

func serializeStringList(v []string, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, s := range v {
		a.Value().String(s)
	}
}

func serializeStringMap(v map[string]string, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	for k, s := range v {
		o.Key(k).String(s)
	}
}

func serializeEnvironment(v []awstypes.KeyValuePair, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, variable := range v {
		o := a.Value().Object()
		if variable.Name != nil {
			o.Key(names.AttrName).String(*variable.Name)
		}
		if variable.Value != nil {
			o.Key(names.AttrValue).String(*variable.Value)
		}
		o.Close()
	}
}

func serializeHealthCheck(v *awstypes.HealthCheck, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Command != nil {
		serializeStringList(v.Command, o.Key("command"))
	}
	if v.Interval != nil {
		o.Key(names.AttrInterval).Integer(*v.Interval)
	}
	if v.Retries != nil {
		o.Key("retries").Integer(*v.Retries)
	}
	if v.StartPeriod != nil {
		o.Key("startPeriod").Integer(*v.StartPeriod)
	}
	if v.Timeout != nil {
		o.Key(names.AttrTimeout).Integer(*v.Timeout)
	}
}

func serializeLinuxParameters(v *awstypes.LinuxParameters, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Capabilities != nil {
		c := o.Key("capabilities").Object()
		if v.Capabilities.Add != nil {
			serializeStringList(v.Capabilities.Add, c.Key("add"))
		}
		if v.Capabilities.Drop != nil {
			serializeStringList(v.Capabilities.Drop, c.Key("drop"))
		}
		c.Close()
	}
	if v.Devices != nil {
		a := o.Key("devices").Array()
		for _, device := range v.Devices {
			d := a.Value().Object()
			if device.ContainerPath != nil {
				d.Key("containerPath").String(*device.ContainerPath)
			}
			if device.HostPath != nil {
				d.Key("hostPath").String(*device.HostPath)
			}
			if device.Permissions != nil {
				p := d.Key(names.AttrPermissions).Array()
				for _, permission := range device.Permissions {
					p.Value().String(string(permission))
				}
				p.Close()
			}
			d.Close()
		}
		a.Close()
	}
	if v.InitProcessEnabled != nil {
		o.Key("initProcessEnabled").Boolean(*v.InitProcessEnabled)
	}
	if v.MaxSwap != nil {
		o.Key("maxSwap").Integer(*v.MaxSwap)
	}
	if v.SharedMemorySize != nil {
		o.Key("sharedMemorySize").Integer(*v.SharedMemorySize)
	}
	if v.Swappiness != nil {
		o.Key("swappiness").Integer(*v.Swappiness)
	}
	if v.Tmpfs != nil {
		a := o.Key("tmpfs").Array()
		for _, tmpfs := range v.Tmpfs {
			t := a.Value().Object()
			if tmpfs.ContainerPath != nil {
				t.Key("containerPath").String(*tmpfs.ContainerPath)
			}
			if tmpfs.MountOptions != nil {
				serializeStringList(tmpfs.MountOptions, t.Key("mountOptions"))
			}
			t.Key(names.AttrSize).Integer(tmpfs.Size)
			t.Close()
		}
		a.Close()
	}
}

func serializeLogConfiguration(v *awstypes.LogConfiguration, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.LogDriver != "" {
		o.Key("logDriver").String(string(v.LogDriver))
	}
	if v.Options != nil {
		serializeStringMap(v.Options, o.Key("options"))
	}
	if v.SecretOptions != nil {
		serializeSecrets(v.SecretOptions, o.Key("secretOptions"))
	}
}

func serializeMountPoints(v []awstypes.MountPoint, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, mount := range v {
		o := a.Value().Object()
		if mount.ContainerPath != nil {
			o.Key("containerPath").String(*mount.ContainerPath)
		}
		if mount.ReadOnly != nil {
			o.Key("readOnly").Boolean(*mount.ReadOnly)
		}
		if mount.SourceVolume != nil {
			o.Key("sourceVolume").String(*mount.SourceVolume)
		}
		o.Close()
	}
}

func serializePortMappings(v []awstypes.PortMapping, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, port := range v {
		o := a.Value().Object()
		if port.AppProtocol != "" {
			o.Key("appProtocol").String(string(port.AppProtocol))
		}
		if port.ContainerPort != nil {
			o.Key("containerPort").Integer(*port.ContainerPort)
		}
		if port.ContainerPortRange != nil {
			o.Key("containerPortRange").String(*port.ContainerPortRange)
		}
		if port.HostPort != nil {
			o.Key("hostPort").Integer(*port.HostPort)
		}
		if port.Name != nil {
			o.Key(names.AttrName).String(*port.Name)
		}
		if port.Protocol != "" {
			o.Key(names.AttrProtocol).String(string(port.Protocol))
		}
		o.Close()
	}
}

func serializeContainerRestartPolicy(v *awstypes.ContainerRestartPolicy, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Enabled != nil {
		o.Key(names.AttrEnabled).Boolean(*v.Enabled)
	}
	if v.IgnoredExitCodes != nil {
		a := o.Key("ignoredExitCodes").Array()
		for _, code := range v.IgnoredExitCodes {
			a.Value().Integer(code)
		}
		a.Close()
	}
	if v.RestartAttemptPeriod != nil {
		o.Key("restartAttemptPeriod").Integer(*v.RestartAttemptPeriod)
	}
}

func serializeSecrets(v []awstypes.Secret, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, secret := range v {
		o := a.Value().Object()
		if secret.Name != nil {
			o.Key(names.AttrName).String(*secret.Name)
		}
		if secret.ValueFrom != nil {
			o.Key("valueFrom").String(*secret.ValueFrom)
		}
		o.Close()
	}
}

func flattenContainerDefinitions(apiObjects []awstypes.ContainerDefinition) (string, error) {
	jsonEncoder := smithyjson.NewEncoder()
	err := serializeContainerDefinitions(apiObjects, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}

func expandContainerDefinitions(tfString string) ([]awstypes.ContainerDefinition, error) {
	var apiObjects []awstypes.ContainerDefinition

	if err := tfjson.DecodeFromString(tfString, &apiObjects); err != nil {
		return nil, err
	}

	for i, apiObject := range apiObjects {
		if inttypes.IsZero(&apiObject) {
			return nil, fmt.Errorf("invalid container definition supplied at index (%d)", i)
		}
		if !isValidVersionConsistency(apiObject) {
			return nil, fmt.Errorf("invalid version consistency value (%[1]s) for container definition supplied at index (%[2]d)", apiObject.VersionConsistency, i)
		}
	}

	containerDefinitions(apiObjects).compactArrays()

	return apiObjects, nil
}

func isValidVersionConsistency(cd awstypes.ContainerDefinition) bool {
	if cd.VersionConsistency == "" {
		return true
	}

	return slices.Contains(enum.EnumValues[awstypes.VersionConsistency](), cd.VersionConsistency)
}
