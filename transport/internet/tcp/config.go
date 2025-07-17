package tcp

import (
	"github.com/meichuanneiku/xray-core/common"
	"github.com/meichuanneiku/xray-core/transport/internet"
)

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
