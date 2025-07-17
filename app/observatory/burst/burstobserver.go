package burst

import (
	"context"
	"encoding/json"
	"log"

	"sync"

	"github.com/meichuanneiku/xray-core/app/observatory"
	"github.com/meichuanneiku/xray-core/common"
	"github.com/meichuanneiku/xray-core/common/errors"
	"github.com/meichuanneiku/xray-core/common/signal/done"
	"github.com/meichuanneiku/xray-core/core"
	"github.com/meichuanneiku/xray-core/features/extension"
	"github.com/meichuanneiku/xray-core/features/outbound"
	"github.com/meichuanneiku/xray-core/features/routing"
	"google.golang.org/protobuf/proto"
)

type Observer struct {
	config *Config
	ctx    context.Context

	statusLock sync.Mutex
	hp         *HealthPing

	finished *done.Instance

	ohm outbound.Manager
}

func (o *Observer) GetObservation(ctx context.Context) (proto.Message, error) {
	return &observatory.ObservationResult{Status: o.createResult()}, nil
}

func (o *Observer) createResult() []*observatory.OutboundStatus {
	var result []*observatory.OutboundStatus
	o.hp.access.Lock()
	defer o.hp.access.Unlock()
	for name, value := range o.hp.Results {
		status := observatory.OutboundStatus{
			Alive:           value.getStatistics().All != value.getStatistics().Fail,
			Delay:           value.getStatistics().Average.Milliseconds(),
			LastErrorReason: "",
			OutboundTag:     name,
			LastSeenTime:    0,
			LastTryTime:     0,
			HealthPing: &observatory.HealthPingMeasurementResult{
				All:       int64(value.getStatistics().All),
				Fail:      int64(value.getStatistics().Fail),
				Deviation: int64(value.getStatistics().Deviation),
				Average:   int64(value.getStatistics().Average),
				Max:       int64(value.getStatistics().Max),
				Min:       int64(value.getStatistics().Min),
			},
		}
		result = append(result, &status)
	}
	return result
}

func (o *Observer) Type() interface{} {
	return extension.ObservatoryType()
}

func (o *Observer) Start() error {
	if o.config != nil && len(o.config.SubjectSelector) != 0 {
		o.finished = done.New()
		o.hp.StartScheduler(func() ([]string, error) {
			hs, ok := o.ohm.(outbound.HandlerSelector)
			if !ok {

				return nil, errors.New("outbound.Manager is not a HandlerSelector")
			}

			outbounds := hs.Select(o.config.SubjectSelector)
			return outbounds, nil
		})
	}
	return nil
}

func (o *Observer) Close() error {
	if o.finished != nil {
		o.hp.StopScheduler()
		return o.finished.Close()
	}
	return nil
}

func New(ctx context.Context, config *Config) (*Observer, error) {
	var outboundManager outbound.Manager
	var dispatcher routing.Dispatcher
	err := core.RequireFeatures(ctx, func(om outbound.Manager, rd routing.Dispatcher) {
		outboundManager = om
		dispatcher = rd
	})
	if err != nil {
		return nil, errors.New("Cannot get depended features").Base(err)
	}
	hp := NewHealthPing(ctx, dispatcher, config.PingConfig)
	return &Observer{
		config: config,
		ctx:    ctx,
		ohm:    outboundManager,
		hp:     hp,
	}, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}

func (o *Observer) AddSelector(tag string) error {
	o.statusLock.Lock()
	defer o.statusLock.Unlock()

	o.config.SubjectSelector = append(o.config.SubjectSelector, tag)

	return o.Start()
}
func (o *Observer) RemoveSelector(tag string) error {
	o.statusLock.Lock()
	defer o.statusLock.Unlock()

	if tag == "" {
		return errors.New("empty tag")
	}
	for i, selector := range o.config.SubjectSelector {
		if selector == tag {
			o.config.SubjectSelector = append(o.config.SubjectSelector[:i], o.config.SubjectSelector[i+1:]...)
			return nil
		}
	}
	return errors.New("tag not found")
}

func (o *Observer) GetConfig(ctx context.Context) string {
	return o.config.String()
}

func (o *Observer) UpdateOtherConfig(config []byte) error {
	o.statusLock.Lock()
	defer o.statusLock.Unlock()

	observatoryConfig := &Config{}
	if err := json.Unmarshal(config, observatoryConfig); err != nil {
		log.Panicf("Failed to unmarshal Routing config: %s", err)
	}

	o.config.PingConfig.Destination = observatoryConfig.PingConfig.Destination
	o.config.PingConfig.Interval = int64(observatoryConfig.PingConfig.Interval)
	o.config.PingConfig.Connectivity = observatoryConfig.PingConfig.Connectivity
	o.config.PingConfig.Timeout = int64(observatoryConfig.PingConfig.Timeout)
	o.config.PingConfig.SamplingCount = int32(observatoryConfig.PingConfig.SamplingCount)

	return nil
}

func (o *Observer) UpdateOtherConfig2(config proto.Message) error {

	config2 := config.(*Config)
	o.config.PingConfig = config2.PingConfig

	return nil
}
