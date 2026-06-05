package topics

import (
	"sync"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
)

type Topics struct {
	registeredTopics  map[string]bool
	publishableTopics map[string]bool
	rpcRequestTopics  map[string]bool
	listeningTopics   map[string]interface{}

	// programmatic registrations (merged with config-based ones at lookup time)
	programmaticSubscribedTopics    map[string]interface{}
	programmaticRpcSubscribedTopics map[string]interface{}
}

var getInstance = sync.OnceValue(func() *Topics {
	instance := &Topics{}
	instance.registeredTopics = make(map[string]bool)
	instance.publishableTopics = make(map[string]bool)
	instance.listeningTopics = make(map[string]interface{})
	instance.rpcRequestTopics = make(map[string]bool)
	instance.programmaticSubscribedTopics = make(map[string]interface{})
	instance.programmaticRpcSubscribedTopics = make(map[string]interface{})
	return instance
})

// GetInstance returns the singleton Topics instance.
func GetInstance() *Topics {
	return getInstance()
}

func (t *Topics) Register() {
	t.registerTopics()
	t.registerPublishingTopics()
	t.registerRpcRequestTopics()
}

func (t *Topics) RegisterManagerTopics(topicName string) {
	t.registeredTopics[configmanager.GetInstance().DeploymentEnv+topicName] = true
	t.publishableTopics[configmanager.GetInstance().DeploymentEnv+topicName] = true
}

// AddRegisteredTopic registers a topic programmatically so it is treated as valid.
// The deployment-env prefix is applied automatically.
func (t *Topics) AddRegisteredTopic(name string) {
	t.registeredTopics[configmanager.GetInstance().DeploymentEnv+name] = true
}

// AddPublishableTopic registers a topic programmatically as both valid and publishable.
func (t *Topics) AddPublishableTopic(name string) {
	t.registeredTopics[configmanager.GetInstance().DeploymentEnv+name] = true
	t.publishableTopics[configmanager.GetInstance().DeploymentEnv+name] = true
}

// AddSubscribedTopic registers a topic programmatically for queue subscription.
// handlerMethod is the method name on the SubscriptionInterface that handles messages.
func (t *Topics) AddSubscribedTopic(name, handlerMethod string) {
	t.registeredTopics[configmanager.GetInstance().DeploymentEnv+name] = true
	t.programmaticSubscribedTopics[configmanager.GetInstance().DeploymentEnv+name] = handlerMethod
}

// AddRpcRequestTopic registers a topic programmatically as an RPC request topic.
func (t *Topics) AddRpcRequestTopic(name string) {
	t.registeredTopics[configmanager.GetInstance().DeploymentEnv+name] = true
	t.rpcRequestTopics[configmanager.GetInstance().DeploymentEnv+name] = true
}

// AddRpcSubscribedTopic registers a topic programmatically for RPC queue subscription.
// handlerMethod is the method name on the SubscriptionInterface that handles messages.
func (t *Topics) AddRpcSubscribedTopic(name, handlerMethod string) {
	t.registeredTopics[configmanager.GetInstance().DeploymentEnv+name] = true
	t.programmaticRpcSubscribedTopics[configmanager.GetInstance().DeploymentEnv+name] = handlerMethod
}

func (t *Topics) registerTopics() {
	registerdTopics := configmanager.GetInstance().RegisteredTopics

	if registerdTopics != nil {

		for i := range registerdTopics {
			t.registeredTopics[configmanager.GetInstance().DeploymentEnv+registerdTopics[i]] = true
		}

	} else {
		if configmanager.GetInstance().PrintWarning {
			logger.Log().Warn("No Register Topics found")
		}
	}
}

func (t *Topics) GetSubscribedTopics() map[string]interface{} {
	merged := make(map[string]interface{})

	// config-based subscriptions
	for key, element := range configmanager.GetInstance().SubscribedTopics {
		merged[configmanager.GetInstance().DeploymentEnv+key] = element
	}

	// programmatic subscriptions (override config if same key)
	for key, element := range t.programmaticSubscribedTopics {
		merged[key] = element
	}

	t.listeningTopics = merged
	return t.listeningTopics
}

func (t *Topics) GetRPCSubcribedTopics() map[string]interface{} {
	merged := make(map[string]interface{})

	// config-based RPC subscriptions
	for key, element := range configmanager.GetInstance().RpcSubscribedTopics {
		merged[configmanager.GetInstance().DeploymentEnv+key] = element
	}

	// programmatic RPC subscriptions (override config if same key)
	for key, element := range t.programmaticRpcSubscribedTopics {
		merged[key] = element
	}

	return merged
}

func (t *Topics) GetNonQueueTopics() map[string]interface{} {
	nonQueueTopics := make(map[string]interface{})
	for key, element := range configmanager.GetInstance().NonQueueSubscribedTopics {
		nonQueueTopics[configmanager.GetInstance().DeploymentEnv+key] = element
	}
	return nonQueueTopics
}

func (t *Topics) registerPublishingTopics() {
	publishingTopics := configmanager.GetInstance().PublishingTopics

	for i := range publishingTopics {
		t.publishableTopics[configmanager.GetInstance().DeploymentEnv+publishingTopics[i]] = true
	}

}

func (t *Topics) registerRpcRequestTopics() {
	requestTopics := configmanager.GetInstance().RpcRequestTopics

	for i := range requestTopics {
		t.rpcRequestTopics[configmanager.GetInstance().DeploymentEnv+requestTopics[i]] = true
	}

}

func (t *Topics) ValidatePublishableTopics(key string) bool {
	if t.ValidateTopic(key) {
		if _, ok := t.publishableTopics[key]; ok {
			return true
		}
	}
	return false
}

func (t *Topics) ValidateRequestTopics(key string) bool {
	if t.ValidateTopic(key) {
		if _, ok := t.rpcRequestTopics[key]; ok {
			return true
		}
	}
	return false
}

func (t *Topics) ValidateTopic(key string) bool {
	if _, ok := t.registeredTopics[key]; ok {
		return true
	}
	return false
}
