package filters

import (
	"encoding/json"
	"go-connector/libs/package_coder"
	"go-connector/libs/pomelo_coder"
)

func NoRouteFilter(route string, pkgId uint64, compressRoute, compressGzip bool) package_coder.BackendMsg {
	return package_coder.BackendMsg{
		Route:         route,
		Payload:       json.RawMessage(`{ "code": 500, "message": null, "data": null, "msg": { "code": 500 } }`),
		PkgID:         pkgId,
		MType:         pomelo_coder.Message["TYPE_RESPONSE"], // 请求类型  TYPE_REQUEST, TYPE_NOTIFY, ...
		CompressRoute: compressRoute,
		CompressGzip:  compressGzip,
	}
}
