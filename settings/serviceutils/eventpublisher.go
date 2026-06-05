package serviceutils

import (
	"math"
	"strings"
	"sync"
	"time"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/internal/topics"
	"github.com/glodb/keel/utils"
	"github.com/glodb/keel/settings/utilsdatatypes"

	"github.com/bytedance/sonic"
	"github.com/rs/xid"
)

type FailSave struct {
	msg Message
	key string
}

type EventPublisher struct {
	ticker        *time.Ticker
	eventMutex    sync.Mutex
	failedMutex   sync.Mutex
	channelsQueue map[string]*utilsdatatypes.Queue
	failedQueue   utilsdatatypes.Queue
}

func (ts *EventPublisher) New() {
	ts.failedQueue.New()
	ts.channelsQueue = make(map[string]*utilsdatatypes.Queue)
	ts.startTimer()
}

func (ts *EventPublisher) startTimer() {
	logger.Log().Debug("Starting Timer", logger.IntField("milli_seconds", configmanager.GetInstance().MessageSendingMilliSeconds))
	ts.ticker = time.NewTicker(time.Duration(configmanager.GetInstance().MessageSendingMilliSeconds) * time.Millisecond)
	go func() {
		for range ts.ticker.C {
			go ts.sendEvents()
		}
	}()
}

func (ts *EventPublisher) StopTimer() {
	ts.ticker.Stop()
	ts.sendEvents()
}

func (ts *EventPublisher) clearFailedQueue() {
	ts.failedMutex.Lock()
	tempFailedQueue := ts.failedQueue.Copy()
	ts.failedQueue.New()
	ts.failedMutex.Unlock()

	for i := range tempFailedQueue {
		faileSave := tempFailedQueue[i].(FailSave)
		if err := GetInstance().GetNats().Publish(faileSave.key, faileSave.msg.Body); err != nil {
			ts.failedMutex.Lock()
			ts.failedQueue.Enqueue(FailSave{msg: faileSave.msg, key: faileSave.key})
			ts.failedMutex.Unlock()
			// logger.Log().Error("Publish Error", logger.StringField("key", faileSave.key), logger.ErrorField("error", err))
		}
	}
}

func (ts *EventPublisher) sendBatches(element []interface{}, dedupId string, key string) {

	logger.Log().Debug("Sending Batches", logger.StringField("key", key), logger.AnyField("data", element))
	length := len(element)
	batches := math.Ceil(float64(length) / float64(configmanager.GetInstance().PublisherBatchSize))
	values := strings.Split(key, ":")
	for i := 0; i < int(batches); i++ {
		batchStart := i * int(configmanager.GetInstance().PublisherBatchSize)
		batchEnd := (i + 1) * int(configmanager.GetInstance().PublisherBatchSize)
		if batchEnd > length {
			batchEnd = length
		}
		body, _ := sonic.Marshal(element[batchStart:batchEnd])
		msg := &Message{
			Header: map[string]string{
				"id":      values[0],
				"dedupid": dedupId,
				"groupid": configmanager.GetInstance().MicroServiceName,
			},
			Body: body,
		}

		logger.Log().Debug("Publishing Message", logger.StringField("key", values[1]), logger.IntField("batch_start", batchStart), logger.IntField("batch_end", batchEnd), logger.AnyField("data", element))

		if err := GetInstance().GetNats().Publish(values[1], msg.Body); err != nil {
			ts.failedMutex.Lock()
			ts.failedQueue.Enqueue(FailSave{msg: *msg, key: values[1]})
			ts.failedMutex.Unlock()
			// logger.Log().Error("Publish Error", logger.StringField("key", values[0]), logger.ErrorField("error", err), logger.IntField("batch_start", batchStart), logger.IntField("batch_end", batchEnd), logger.AnyField("data", element))
		}
	}
}

func (ts *EventPublisher) sendEvents() {

	ts.clearFailedQueue()

	ts.eventMutex.Lock()
	newMap := utils.GetInstance().CopyMap(ts.channelsQueue)
	ts.channelsQueue = make(map[string]*utilsdatatypes.Queue)
	ts.eventMutex.Unlock()

	for key, element := range newMap {
		logger.Log().Debug("Sending Events", logger.StringField("key", key), logger.AnyField("data", element))
		dedupId := xid.New().String()
		ts.sendBatches(element, dedupId, key)
	}
}

func (ts *EventPublisher) publishEvent(data interface{}, serviceName string, topic string) error {
	// Marshal to JSON string
	if topics.GetInstance().ValidatePublishableTopics(topic) {

		logger.Log().Debug("Publishing Event", logger.StringField("topic", topic), logger.StringField("service_name", serviceName), logger.AnyField("data", data))

		key := serviceName + ":" + topic
		ts.eventMutex.Lock()
		defer ts.eventMutex.Unlock()
		_, ok := ts.channelsQueue[key]

		if !ok {
			ts.channelsQueue[key] = &utilsdatatypes.Queue{}
			ts.channelsQueue[key].New()
			ts.channelsQueue[key].Enqueue(data)
		} else {
			ts.channelsQueue[key].Enqueue(data)
		}
	} else {
		// logger.Log().Error("Topic not registered", logger.StringField("service_name", serviceName), logger.StringField("topic", topic))
	}
	return nil
}
