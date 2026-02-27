package webhook

type RecommendedValue struct {
	ResourceRequest struct {
		Containers []struct {
			ContainerName string `yaml:"containerName"`
			Target        struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"target"`
		} `yaml:"containers"`
	} `yaml:"resourceRequest"`
}

type JSONPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
