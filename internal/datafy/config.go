package datafy

const DefaultUrl = "https://iac.datafy.io"

type Config struct {
	Token string
	Url   string
}

// Copy will return a shallow copy of the Config object.
func (c Config) Copy() Config {
	cp := c
	return cp
}
