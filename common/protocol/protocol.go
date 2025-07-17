package protocol // import "github.com/meichuanneiku/xray-core/common/protocol"

import (
	"errors"
)

var ErrProtoNeedMoreData = errors.New("protocol matches, but need more data to complete sniffing")
