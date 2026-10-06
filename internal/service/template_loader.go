package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadTemplates(dir string) (map[string]Template, error) {
	templates := make(map[string]Template)

	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("find templates: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read template %q: %w", file, err)
		}

		var template Template

		if err := yaml.Unmarshal(data, &template); err != nil {
			return nil, fmt.Errorf("unmarshal template %q: %w", file, err)
		}

		templates[template.Name] = template
	}

	log.Printf("%d templates loaded", len(templates))

	return templates, nil
}
