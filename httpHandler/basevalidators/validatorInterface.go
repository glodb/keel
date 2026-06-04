package basevalidators

type ValidatorInterface interface {
	GetRules(apiName string) map[string]interface{}
}
