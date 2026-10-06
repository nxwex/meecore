package server

type (
	CreateContainerRequest struct {
		Name     string            `json:"name"`
		Template string            `json:"template"`
		Image    string            `json:"image"`
		Env      map[string]string `json:"env"`
	}

	CreateInstanceRequest struct {
		NodeID   int64  `json:"node_id"`
		Name     string `json:"name"`
		Template string `json:"template"`
	}
)
