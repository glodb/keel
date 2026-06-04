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
}

var getInstance = sync.OnceValue(func() *Topics {
	instance := &Topics{}
	instance.registeredTopics = make(map[string]bool)
	instance.publishableTopics = make(map[string]bool)
	instance.listeningTopics = make(map[string]interface{})
	instance.rpcRequestTopics = make(map[string]bool)
	return instance
})

// Singleton. Returns a single object of Topic
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
	subscribedTopics := configmanager.GetInstance().SubscribedTopics
	newSub := make(map[string]interface{})

	if subscribedTopics == nil {
		return newSub
	}

	for key, element := range subscribedTopics {
		newSub[configmanager.GetInstance().DeploymentEnv+key] = element
	}
	if subscribedTopics != nil {
		t.listeningTopics = newSub
	} else {
		return nil
	}
	return t.listeningTopics
}

func (t *Topics) GetRPCSubcribedTopics() map[string]interface{} {
	rpcListeningTopics := make(map[string]interface{})
	for key, element := range configmanager.GetInstance().RpcSubscribedTopics {
		rpcListeningTopics[configmanager.GetInstance().DeploymentEnv+key] = element
	}
	return rpcListeningTopics
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
