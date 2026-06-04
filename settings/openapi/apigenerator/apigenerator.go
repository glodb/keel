package apigenerator

// RegisterDocumentAPI registers an API endpoint into the OpenAPI document generator.
//
// NOTE: This repository currently does not generate OpenAPI docs at runtime in production builds.
// This stub exists to keep builds green for packages that reference this hook.
func RegisterDocumentAPI(path, method, summary, description string, tags []string, requestBody, responses map[string]interface{}) error {
	return nil
}
