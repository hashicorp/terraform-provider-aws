package tfaddr

import (
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
)

// A ModulePackage is an extra indirection over a ModulePackage where
// we use a module registry to translate a more symbolic address (and
// associated version constraint given out of band) into a physical source
// location.
//
// ModulePackage is distinct from ModulePackage because they have
// disjoint use-cases: registry package addresses are only used to query a
// registry in order to find a real module package address. These being
// distinct is intended to help future maintainers more easily follow the
// series of steps in the module installer, with the help of the type checker.
type ModulePackage struct {
	Host         svchost.Hostname
	Namespace    string
	Name         string
	TargetSystem string
}

func (s ModulePackage) String() string {
	// Note: we're using the "display" form of the hostname here because
	// for our service hostnames "for display" means something different:
	// it means to render non-ASCII characters directly as Unicode
	// characters, rather than using the "punycode" representation we
	// use for internal processing, and so the "display" representation
	// is actually what users would write in their configurations.
	return s.Host.ForDisplay() + "/" + s.ForRegistryProtocol()
}

func (s ModulePackage) ForDisplay() string {
	if s.Host == DefaultModuleRegistryHost {
		return s.ForRegistryProtocol()
	}
	return s.Host.ForDisplay() + "/" + s.ForRegistryProtocol()
}

// ForRegistryProtocol returns a string representation of just the namespace,
// name, and target system portions of the address, always omitting the
// registry hostname and the subdirectory portion, if any.
//
// This is primarily intended for generating addresses to send to the
// registry in question via the registry protocol, since the protocol
// skips sending the registry its own hostname as part of identifiers.
func (s ModulePackage) ForRegistryProtocol() string {
	var buf strings.Builder
	buf.WriteString(s.Namespace)
	buf.WriteByte('/')
	buf.WriteString(s.Name)
	buf.WriteByte('/')
	buf.WriteString(s.TargetSystem)
	return buf.String()
}
