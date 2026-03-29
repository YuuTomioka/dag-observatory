package spec

type WorkflowSpec struct {
	Name    string     `yaml:"name"`
	Version string     `yaml:"version"`
	Inputs  []string   `yaml:"inputs"`
	Nodes   []NodeSpec `yaml:"nodes"`
}

type NodeSpec struct {
	ID     string         `yaml:"id"`
	Kind   string         `yaml:"kind"`
	Config map[string]any `yaml:"config"`
}
