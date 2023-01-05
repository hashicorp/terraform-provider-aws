package actionlint

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:generate go run ./scripts/generate-popular-actions -s remote -f go ./popular_actions.go

// ActionMetadataInput is input metadata in "inputs" section in action.yml metadata file.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
type ActionMetadataInput struct {
	// Name is a name of this input.
	Name string `json:"name"`
	// Required is true when this input is mandatory to run the action.
	Required bool `json:"required"`
}

// ActionMetadataInputs is a map from input ID to its metadata. Keys are in lower case since input
// names are case-insensitive.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
type ActionMetadataInputs map[string]*ActionMetadataInput

// UnmarshalYAML implements yaml.Unmarshaler.
func (inputs *ActionMetadataInputs) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return expectedMapping("inputs", n)
	}

	type actionInputMetadata struct {
		Required bool    `yaml:"required"`
		Default  *string `yaml:"default"`
	}

	md := make(ActionMetadataInputs, len(n.Content)/2)
	for i := 0; i < len(n.Content); i += 2 {
		k, v := n.Content[i].Value, n.Content[i+1]

		var m actionInputMetadata
		if err := v.Decode(&m); err != nil {
			return err
		}

		id := strings.ToLower(k)
		if _, ok := md[id]; ok {
			return fmt.Errorf("input %q is duplicated", k)
		}

		md[id] = &ActionMetadataInput{k, m.Required && m.Default == nil}
	}

	*inputs = md
	return nil
}

// ActionMetadataOutput is output metadata in "outputs" section in action.yml metadata file.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#outputs-for-composite-actions
type ActionMetadataOutput struct {
	Name string `json:"name"`
}

// ActionMetadataOutputs is a map from output ID to its metadata. Keys are in lower case since output
// names are case-insensitive.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#outputs-for-composite-actions
type ActionMetadataOutputs map[string]*ActionMetadataOutput

// UnmarshalYAML implements yaml.Unmarshaler.
func (inputs *ActionMetadataOutputs) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return expectedMapping("outputs", n)
	}

	md := make(ActionMetadataOutputs, len(n.Content)/2)
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i].Value
		id := strings.ToLower(k)
		if _, ok := md[id]; ok {
			return fmt.Errorf("output %q is duplicated", k)
		}
		md[id] = &ActionMetadataOutput{k}
	}

	*inputs = md
	return nil
}

// ActionMetadata represents structure of action.yaml.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
type ActionMetadata struct {
	// Name is "name" field of action.yaml
	Name string `yaml:"name" json:"name"`
	// Inputs is "inputs" field of action.yaml
	Inputs ActionMetadataInputs `yaml:"inputs" json:"inputs"`
	// Outputs is "outputs" field of action.yaml. Key is name of output. Description is omitted
	// since actionlint does not use it.
	Outputs ActionMetadataOutputs `yaml:"outputs" json:"outputs"`
	// SkipInputs is flag to specify behavior of inputs check. When it is true, inputs for this
	// action will not be checked.
	SkipInputs bool `json:"skip_inputs"`
	// SkipOutputs is flag to specify a bit loose typing to outputs object. If it is set to
	// true, the outputs object accepts any properties along with strictly typed props.
	SkipOutputs bool `json:"skip_outputs"`
}

// LocalActionsCache is cache for local actions' metadata. It avoids repeating to find/read/parse
// local action's metadata file (action.yml).
// This cache is not available across multiple repositories. One LocalActionsCache instance needs
// to be created per one repository.
type LocalActionsCache struct {
	mu    sync.RWMutex
	proj  *Project // might be nil
	cache map[string]*ActionMetadata
	dbg   io.Writer
}

// NewLocalActionsCache creates new LocalActionsCache instance for the given project.
func NewLocalActionsCache(proj *Project, dbg io.Writer) *LocalActionsCache {
	return &LocalActionsCache{
		proj:  proj,
		cache: map[string]*ActionMetadata{},
		dbg:   dbg,
	}
}

func (c *LocalActionsCache) debug(format string, args ...interface{}) {
	if c.dbg == nil {
		return
	}
	format = "[LocalActionsCache] " + format + "\n"
	fmt.Fprintf(c.dbg, format, args...)
}

func (c *LocalActionsCache) readCache(key string) (*ActionMetadata, bool) {
	c.mu.RLock()
	m, ok := c.cache[key]
	c.mu.RUnlock()
	return m, ok
}

func (c *LocalActionsCache) writeCache(key string, val *ActionMetadata) {
	c.mu.Lock()
	c.cache[key] = val
	c.mu.Unlock()
}

// FindMetadata finds metadata for given spec. The spec should indicate for local action hence it
// should start with "./". The first return value can be nil even if error did not occur.
// LocalActionCache caches that the action was not found. At first search, it returns an error that
// the action was not found. But at the second search, it does not return an error even if the result
// is nil. This behavior prevents repeating to report the same error from multiple places.
// Calling this method is thread-safe.
func (c *LocalActionsCache) FindMetadata(spec string) (*ActionMetadata, error) {
	if c.proj == nil || !strings.HasPrefix(spec, "./") {
		return nil, nil
	}

	if m, ok := c.readCache(spec); ok {
		c.debug("Cache hit for %s: %v", spec, m)
		return m, nil
	}

	dir := filepath.Join(c.proj.RootDir(), filepath.FromSlash(spec))
	b, ok := c.readLocalActionMetadataFile(dir)
	if !ok {
		c.debug("No action metadata found in %s", dir)
		// Remember action was not found
		c.writeCache(spec, nil)
		// Do not complain about the action does not exist (#25, #40).
		// It seems a common pattern that the local action does not exist in the repository
		// (e.g. Git submodule) and it is cloned at running workflow (due to a private repository).
		return nil, nil
	}

	var meta ActionMetadata
	if err := yaml.Unmarshal(b, &meta); err != nil {
		c.writeCache(spec, nil) // Remember action was invalid
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		return nil, fmt.Errorf("could not parse action metadata in %q: %s", dir, msg)
	}

	c.debug("New metadata parsed from action %s: %v", dir, &meta)

	c.writeCache(spec, &meta)
	return &meta, nil
}

func (c *LocalActionsCache) readLocalActionMetadataFile(dir string) ([]byte, bool) {
	for _, p := range []string{
		filepath.Join(dir, "action.yaml"),
		filepath.Join(dir, "action.yml"),
	} {
		if b, err := os.ReadFile(p); err == nil {
			return b, true
		}
	}

	return nil, false
}

// LocalActionsCacheFactory is a factory to create LocalActionsCache instances. LocalActionsCache
// should be created for each repositories. LocalActionsCacheFactory creates new LocalActionsCache
// instance per repository (project).
type LocalActionsCacheFactory struct {
	caches map[string]*LocalActionsCache
	dbg    io.Writer
}

// GetCache returns LocalActionsCache instance for the given project. One LocalActionsCache is
// created per one repository. Created instances are cached and will be used when caches are
// requested for the same projects. This method is not thread safe.
func (f *LocalActionsCacheFactory) GetCache(p *Project) *LocalActionsCache {
	r := p.RootDir()
	if c, ok := f.caches[r]; ok {
		return c
	}
	c := NewLocalActionsCache(p, f.dbg)
	f.caches[r] = c
	return c
}

// NewLocalActionsCacheFactory creates a new LocalActionsCacheFactory instance.
func NewLocalActionsCacheFactory(dbg io.Writer) *LocalActionsCacheFactory {
	return &LocalActionsCacheFactory{map[string]*LocalActionsCache{}, dbg}
}
