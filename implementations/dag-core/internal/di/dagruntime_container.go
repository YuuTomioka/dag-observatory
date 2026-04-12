package di

import (
	"context"
	"fmt"
	"strings"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	clockinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/clock"
	eventstoreinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/eventstore"
	observerinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/observer"
	recorderinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/recorder"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
)

type DAGRuntimeContainer struct {
	ArtifactStore *artifactinfra.MemoryStore
	StateStore    *StateStoreBundle
	EventProducer *eventstoreinfra.KafkaProducer
	EventConsumer *eventstoreinfra.KafkaConsumer
	EventStream   port.EventStream
	Observer      *observerinfra.OTelObserver
	Recorder      port.Recorder
	RecorderClose func() error
	ReplayStats   recorderinfra.ReplayStats
	RunReader     port.RunReader
	Runner        *engine.Runner
	Driver        *driver.Driver
	Usecase       *usecase.RunWorkflow
}

type StateStoreBundle struct {
	Store state.Store
	Close func() error
}

func NewDAGRuntimeContainer(cfg Config, otelc *OTelContainer, compiled pipeline.Compiled) (*DAGRuntimeContainer, error) {
	artifactStore := artifactinfra.NewMemoryStore()
	stateStore, err := newStateStore(cfg)
	if err != nil {
		return nil, err
	}
	producer, err := newKafkaProducer(cfg)
	if err != nil {
		return nil, err
	}
	consumer, err := newKafkaConsumer(cfg)
	if err != nil {
		return nil, err
	}
	var eventStream port.EventStream
	if consumer != nil {
		eventStream = eventstoreinfra.NewKafkaEventStream(consumer)
	}
	var observer *observerinfra.OTelObserver
	if otelc != nil {
		observer = observerinfra.NewOTelObserver(otelc.IntentLog, otelc.Metrics)
	} else {
		observer = observerinfra.NewOTelObserver(nil, nil)
	}
	runReader := recorderinfra.NewInMemoryRecorder()
	var recorder port.Recorder = runReader
	recorderClose := func() error { return nil }
	replayStats := recorderinfra.ReplayStats{}
	if cfg.ObservationJSONLPath != "" {
		replayStats, err = recorderinfra.ReplayJSONL(cfg.ObservationJSONLPath, runReader)
		if err != nil {
			return nil, err
		}
		jsonlRecorder, err := recorderinfra.NewJSONLRecorder(cfg.ObservationJSONLPath, runReader)
		if err != nil {
			return nil, err
		}
		recorder = jsonlRecorder
		recorderClose = jsonlRecorder.Close
	}
	observer.OnCompile(context.Background(), port.CompileInfo{
		WorkflowName: compiled.Name,
		NodeCount:    len(compiled.Order),
	})

	runner := &engine.Runner{
		ArtifactStore: artifactStore,
		StateStore:    stateStore.Store,
		Observer:      observer,
		Recorder:      recorder,
		Policy: policy.Policy{
			DefaultRetry: policy.RetryPolicy{MaxAttempts: 1},
		},
	}

	driver := &driver.Driver{
		Runner:   runner,
		Compiled: compiled,
	}

	var enqueuer port.EventEnqueuer
	if producer != nil {
		enqueuer = producer
	}

	uc := &usecase.RunWorkflow{
		Driver:        driver,
		Enqueuer:      enqueuer,
		Clock:         clockinfra.NewRealClock(),
		ProducerTopic: kafkaTaskTopic(cfg),
	}

	return &DAGRuntimeContainer{
		ArtifactStore: artifactStore,
		StateStore:    stateStore,
		EventProducer: producer,
		EventConsumer: consumer,
		EventStream:   eventStream,
		Observer:      observer,
		Recorder:      recorder,
		RecorderClose: recorderClose,
		ReplayStats:   replayStats,
		RunReader:     runReader,
		Runner:        runner,
		Driver:        driver,
		Usecase:       uc,
	}, nil
}

func newStateStore(cfg Config) (*StateStoreBundle, error) {
	switch cfg.StateStoreType {
	case "memory", "":
		store := stateinfra.NewMemoryStore()
		return &StateStoreBundle{
			Store: store,
			Close: func() error { return nil },
		}, nil
	case "bolt":
		if cfg.BoltPath == "" {
			return nil, fmt.Errorf("dagruntime: bolt path is required")
		}
		store, err := stateinfra.NewBoltStore(cfg.BoltPath, nil)
		if err != nil {
			return nil, err
		}
		return &StateStoreBundle{
			Store: store,
			Close: store.Close,
		}, nil
	default:
		return nil, fmt.Errorf("dagruntime: unsupported state store type: %s", cfg.StateStoreType)
	}
}

func newKafkaProducer(cfg Config) (*eventstoreinfra.KafkaProducer, error) {
	if cfg.EventStoreType != "kafka" {
		return nil, nil
	}
	topic := kafkaTaskTopic(cfg)
	if cfg.KafkaBrokers == "" || topic == "" {
		return nil, fmt.Errorf("dagruntime: kafka brokers/topic are required when EVENT_STORE_TYPE=kafka")
	}
	brokers := splitCSV(cfg.KafkaBrokers)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("dagruntime: kafka brokers are required")
	}
	return eventstoreinfra.NewKafkaProducer(brokers, topic)
}

func kafkaTaskTopic(cfg Config) string {
	if cfg.KafkaTaskTopic != "" {
		return cfg.KafkaTaskTopic
	}
	return cfg.KafkaTopic
}

func newKafkaConsumer(cfg Config) (*eventstoreinfra.KafkaConsumer, error) {
	if cfg.EventStoreType != "kafka" {
		return nil, nil
	}
	topic := cfg.KafkaEventTopic
	if topic == "" {
		topic = cfg.KafkaTopic
	}
	if cfg.KafkaBrokers == "" || topic == "" {
		return nil, fmt.Errorf("dagruntime: kafka brokers/topic are required when EVENT_STORE_TYPE=kafka")
	}
	brokers := splitCSV(cfg.KafkaBrokers)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("dagruntime: kafka brokers are required")
	}
	return eventstoreinfra.NewKafkaConsumer(brokers, topic, cfg.KafkaGroupID)
}

func splitCSV(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
