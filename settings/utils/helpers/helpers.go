package helpers

import "sync"

type helpers struct {
}

var getInstance = sync.OnceValue(func() *helpers {
	instance := &helpers{}
	return instance
})

// Singleton. Returns a single object of Utils
func GetInstance() *helpers {
	return getInstance()
}

func (u *helpers) GetStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func (u *helpers) GetIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}

func (u *helpers) GetFlaotValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

func (u *helpers) GetBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}
