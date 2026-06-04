package models

type Service struct {
	Name       string `json:"name"`
	ConfigPath string `json:"configPath"`
}

type Services struct {
	Services []Service `json:"services"`
}

func (s *Services) GetService(name string) *Service {
	for _, service := range s.Services {
		if service.Name == name {
			return &service
		}
	}
	return nil
}

func (s *Services) GetAllServices() []Service {
	return s.Services
}
