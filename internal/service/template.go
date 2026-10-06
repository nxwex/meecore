package service

type Template struct {
	Name        string `json:"name" yaml:"name"`
	DisplayName string `json:"display_name" yaml:"display_name"`
	Image       string `json:"image" yaml:"image"`

	Env     []EnvSetting `json:"env" yaml:"env"`
	Ports   []Port       `json:"ports" yaml:"ports"`
	Volumes []Volume     `json:"volumes" yaml:"volumes"`
}

type EnvSetting struct {
	Name     string `json:"name" yaml:"name"`
	Required bool   `json:"required" yaml:"required"`
	Default  string `json:"default" yaml:"default"`
}

type Port struct {
	Container int    `json:"container" yaml:"container"`
	Protocol  string `json:"protocol" yaml:"protocol"`
}

type Volume struct {
	Container string `json:"container" yaml:"container"`
}
