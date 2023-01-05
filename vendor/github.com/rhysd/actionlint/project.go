package actionlint

import (
	"os"
	"path/filepath"
	"strings"
)

// Project represents one GitHub project. One Git repository corresponds to one project.
type Project struct {
	root   string
	config *Config
}

func absPath(path string) string {
	if p, err := filepath.Abs(path); err == nil {
		path = p
	}
	return path
}

// findProject creates new Project instance by finding a project which the given path belongs to.
// A project must be a Git repository and have ".github/workflows" directory.
func findProject(path string) *Project {
	d := absPath(path)
	for {
		w := filepath.Join(d, ".github", "workflows")
		if s, err := os.Stat(w); err == nil && s.IsDir() {
			g := filepath.Join(d, ".git")
			if _, err := os.Stat(g); err == nil { // Note: .git may be a file
				return &Project{root: d}
			}
		}

		p := filepath.Dir(d)
		if p == d {
			return nil
		}
		d = p
	}
}

// RootDir returns a root directory path of the GitHub project repository.
func (p *Project) RootDir() string {
	return p.root
}

// WorkflowsDir returns a ".github/workflows" directory path of the GitHub project repository.
// This method does not check if the directory exists.
func (p *Project) WorkflowsDir() string {
	return filepath.Join(p.root, ".github", "workflows")
}

// Knows returns true when the project knows the given file. When a file is included in the
// project's directory, the project knows the file.
func (p *Project) Knows(path string) bool {
	// TODO: strings.HasPrefix is not perfect to check file path
	return strings.HasPrefix(absPath(path), p.root)
}

// Config returns config object of the GitHub project repository. The config file is read from
// ".github/actionlint.yaml" or ".github/actionlint.yml".
func (p *Project) Config() (*Config, error) {
	if p.config != nil {
		return p.config, nil
	}

	for _, f := range []string{"actionlint.yaml", "actionlint.yml"} {
		path := filepath.Join(p.root, ".github", f)
		b, err := os.ReadFile(path)
		if err != nil {
			continue // file does not exist
		}
		cfg, err := parseConfig(b, path)
		if err != nil {
			return nil, err
		}
		p.config = cfg
		return cfg, nil
	}

	return nil, nil // not found
}

// Projects represents set of projects. It caches Project instances which was created previously
// and reuses them.
type Projects struct {
	known []*Project
}

// NewProjects creates new Projects instance.
func NewProjects() *Projects {
	return &Projects{}
}

// At returns the Project instance which the path belongs to.
func (ps *Projects) At(path string) *Project {
	for _, p := range ps.known {
		if p.Knows(path) {
			return p
		}
	}

	p := findProject(path)
	if p != nil {
		ps.known = append(ps.known, p)
	}

	return p
}
