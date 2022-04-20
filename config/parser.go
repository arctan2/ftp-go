package config

type ConfigHandler interface {
	GetInitDir() string
	IsRestricted(string) bool
	IsInBounds(string) bool
}

func ParseConfigFile(path string) (ConfigHandler, error) {
	var c = &config{}
	if err := c.parse(path); err != nil {
		return nil, err
	}
	return c, nil
}
