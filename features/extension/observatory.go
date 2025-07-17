package extension

import (
	"context"
	"github.com/meichuanneiku/xray-core/common/serial"

	"github.com/meichuanneiku/xray-core/features"
	"google.golang.org/protobuf/proto"
)

type Observatory interface {
	features.Feature

	GetObservation(ctx context.Context) (proto.Message, error)

	AddSelector(tag string) error
	RemoveSelector(tag string) error
	UpdateOtherConfig(config *serial.TypedMessage) error
	GetConfig(ctx context.Context) string
}

func ObservatoryType() interface{} {
	return (*Observatory)(nil)
}
