package tfaddr

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
)

// Module is representing a module listed in a Terraform module
// registry.
type Module struct {
	// Package is the registry package that the target module belongs to.
	// The module installer must translate this into a ModuleSourceRemote
	// using the registry API and then take that underlying address's
	// Package in order to find the actual package location.
	Package ModulePackage

	// If Subdir is non-empty then it represents a sub-directory within the
	// remote package that the registry address eventually resolves to.
	// This will ultimately become the suffix of the Subdir of the
	// ModuleSourceRemote that the registry address translates to.
	//
	// Subdir uses a normalized forward-slash-based path syntax within the
	// virtual filesystem represented by the final package. It will never
	// include `../` or `./` sequences.
	Subdir string
}

// DefaultModuleRegistryHost is the hostname used for registry-based module
// source addresses that do not have an explicit hostname.
const DefaultModuleRegistryHost = svchost.Hostname("registry.terraform.io")

var moduleRegistryNamePattern = regexp.MustCompile("^[0-9A-Za-z](?:[0-9A-Za-z-_]{0,62}[0-9A-Za-z])?$")
var moduleRegistryTargetSystemPattern = regexp.MustCompile("^[0-9a-z]{1,64}$")

// ParseModuleSource only accepts module registry addresses, and
// will reject any other address type.
func ParseModuleSource(raw string) (Module, error) {
	var err error

	var subDir string
	raw, subDir = splitPackageSubdir(raw)
	if strings.HasPrefix(subDir, "../") {
		return Module{}, fmt.Errorf("subdirectory path %q leads outside of the module package", subDir)
	}

	parts := strings.Split(raw, "/")
	// A valid registry address has either three or four parts, because the
	// leading hostname part is optional.
	if len(parts) != 3 && len(parts) != 4 {
		return Module{}, fmt.Errorf("a module registry source address must have either three or four slash-separated components")
	}

	host := DefaultModuleRegistryHost
	if len(parts) == 4 {
		host, err = svchost.ForComparison(parts[0])
		if err != nil {
			// The svchost library doesn't produce very good error messages to
			// return to an end-user, so we'll use some custom ones here.
			switch {
			case strings.Contains(parts[0], "--"):
				// Looks like possibly punycode, which we don't allow here
				// to ensure that source addresses are written readably.
				return Module{}, fmt.Errorf("invalid module registry hostname %q; internationalized domain names must be given as direct unicode characters, not in punycode", parts[0])
			default:
				return Module{}, fmt.Errorf("invalid module registry hostname %q", parts[0])
			}
		}
		if !strings.Contains(host.String(), ".") {
			return Module{}, fmt.Errorf("invalid module registry hostname: must contain at least one dot")
		}
		// Discard the hostname prefix now that we've processed it
		parts = parts[1:]
	}

	ret := Module{
		Package: ModulePackage{
			Host: host,
		},

		Subdir: subDir,
	}

	if host == svchost.Hostname("github.com") || host == svchost.Hostname("bitbucket.org") {
		return ret, fmt.Errorf("can't use %q as a module registry host, because it's reserved for installing directly from version control repositories", host)
	}

	if ret.Package.Namespace, err = parseModuleRegistryName(parts[0]); err != nil {
		if strings.Contains(parts[0], ".") {
			// Seems like the user omitted one of the latter components in
			// an address with an explicit hostname.
			return ret, fmt.Errorf("source address must have three more components after the hostname: the namespace, the name, and the target system")
		}
		return ret, fmt.Errorf("invalid namespace %q: %s", parts[0], err)
	}
	if ret.Package.Name, err = parseModuleRegistryName(parts[1]); err != nil {
		return ret, fmt.Errorf("invalid module name %q: %s", parts[1], err)
	}
	if ret.Package.TargetSystem, err = parseModuleRegistryTargetSystem(parts[2]); err != nil {
		if strings.Contains(parts[2], "?") {
			// The user was trying to include a query string, probably?
			return ret, fmt.Errorf("module registry addresses may not include a query string portion")
		}
		return ret, fmt.Errorf("invalid target system %q: %s", parts[2], err)
	}

	return ret, nil
}

// MustParseModuleSource is a wrapper around ParseModuleSource that panics if
// it returns an error.
func MustParseModuleSource(raw string) (Module) {
	mod, err := ParseModuleSource(raw)
	if err != nil {
		panic(err)
	}
	return mod
}

// parseModuleRegistryName validates and normalizes a string in either the
// "namespace" or "name" position of a module registry source address.
func parseModuleRegistryName(given string) (string, error) {
	// Similar to the names in provider source addresses, we defined these
	// to be compatible with what filesystems and typical remote systems
	// like GitHub allow in names. Unfortunately we didn't end up defining
	// these exactly equivalently: provider names can only use dashes as
	// punctuation, whereas module names can use underscores. So here we're
	// using some regular expressions from the original module source
	// implementation, rather than using the IDNA rules as we do in
	// ParseProviderPart.

	if !moduleRegistryNamePattern.MatchString(given) {
		return "", fmt.Errorf("must be between one and 64 characters, including ASCII letters, digits, dashes, and underscores, where dashes and underscores may not be the prefix or suffix")
	}

	// We also skip normalizing the name to lowercase, because we historically
	// didn't do that and so existing module registries might be doing
	// case-sensitive matching.
	return given, nil
}

// parseModuleRegistryTargetSystem validates and normalizes a string in the
// "target system" position of a module registry source address. This is
// what we historically called "provider" but never actually enforced as
// being a provider address, and now _cannot_ be a provider address because
// provider addresses have three slash-separated components of their own.
func parseModuleRegistryTargetSystem(given string) (string, error) {
	// Similar to the names in provider source addresses, we defined these
	// to be compatible with what filesystems and typical remote systems
	// like GitHub allow in names. Unfortunately we didn't end up defining
	// these exactly equivalently: provider names can't use dashes or
	// underscores. So here we're using some regular expressions from the
	// original module source implementation, rather than using the IDNA rules
	// as we do in ParseProviderPart.

	if !moduleRegistryTargetSystemPattern.MatchString(given) {
		return "", fmt.Errorf("must be between one and 64 ASCII letters or digits")
	}

	// We also skip normalizing the name to lowercase, because we historically
	// didn't do that and so existing module registries might be doing
	// case-sensitive matching.
	return given, nil
}

// String returns a full representation of the address, including any
// additional components that are typically implied by omission in
// user-written addresses.
//
// We typically use this longer representation in error message, in case
// the inclusion of normally-omitted components is helpful in debugging
// unexpected behavior.
func (s Module) String() string {
	if s.Subdir != "" {
		return s.Package.String() + "//" + s.Subdir
	}
	return s.Package.String()
}

// ForDisplay is similar to String but instead returns a representation of
// the idiomatic way to write the address in configuration, omitting
// components that are commonly just implied in addresses written by
// users.
//
// We typically use this shorter representation in informational messages,
// such as the note that we're about to start downloading a package.
func (s Module) ForDisplay() string {
	if s.Subdir != "" {
		return s.Package.ForDisplay() + "//" + s.Subdir
	}
	return s.Package.ForDisplay()
}

// splitPackageSubdir detects whether the given address string has a
// subdirectory portion, and if so returns a non-empty subDir string
// along with the trimmed package address.
//
// If the given string doesn't have a subdirectory portion then it'll
// just be returned verbatim in packageAddr, with an empty subDir value.
func splitPackageSubdir(given string) (packageAddr, subDir string) {
	packageAddr, subDir = sourceDirSubdir(given)
	if subDir != "" {
		subDir = path.Clean(subDir)
	}
	return packageAddr, subDir
}

// sourceDirSubdir takes a source URL and returns a tuple of the URL without
// the subdir and the subdir.
//
// ex:
//   dom.com/path/?q=p               => dom.com/path/?q=p, ""
//   proto://dom.com/path//*?q=p     => proto://dom.com/path?q=p, "*"
//   proto://dom.com/path//path2?q=p => proto://dom.com/path?q=p, "path2"
func sourceDirSubdir(src string) (string, string) {
	// URL might contains another url in query parameters
	stop := len(src)
	if idx := strings.Index(src, "?"); idx > -1 {
		stop = idx
	}

	// Calculate an offset to avoid accidentally marking the scheme
	// as the dir.
	var offset int
	if idx := strings.Index(src[:stop], "://"); idx > -1 {
		offset = idx + 3
	}

	// First see if we even have an explicit subdir
	idx := strings.Index(src[offset:stop], "//")
	if idx == -1 {
		return src, ""
	}

	idx += offset
	subdir := src[idx+2:]
	src = src[:idx]

	// Next, check if we have query parameters and push them onto the
	// URL.
	if idx = strings.Index(subdir, "?"); idx > -1 {
		query := subdir[idx:]
		subdir = subdir[:idx]
		src += query
	}

	return src, subdir
}
