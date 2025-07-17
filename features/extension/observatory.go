package extension

import (
	"context"

	"github.com/xtls/xray-core/features"
	"google.golang.org/protobuf/proto"
)

type Observatory interface {
	features.Feature

	GetObservation(ctx context.Context) (proto.Message, error)

	AddSelector(tag string) error
	RemoveSelector(tag string) error
	UpdateOtherConfig(config []byte) error
	UpdateOtherConfig2(config proto.Message) error
	GetConfig(ctx context.Context) string
}

func ObservatoryType() interface{} {
	return (*Observatory)(nil)
}
