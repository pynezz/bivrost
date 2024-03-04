package model

// Config is the main struct for the configuration
type Config struct {
	Sources []struct {
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Location string   `json:"location"`
		Format   string   `json:"format"`
		Tags     []string `json:"tags"`
	} `json:"sources"`
}
