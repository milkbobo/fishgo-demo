package middleware

import (
	. "github.com/milkbobo/fishgoweb/app/metric"
	. "github.com/milkbobo/fishgoweb/app/router"
	. "github.com/milkbobo/fishgoweb/encoding"
	"net/http"
	"time"
)

func NewPathMetricMiddleware(metric Metric) RouterMiddleware {
	return func(prev RouterMiddlewareContext) RouterMiddlewareContext {
		pathEncoding, err := EncodeUrl(prev.Data["path"].(string))
		if err != nil {
			panic(err)
		}
		pathRequest := metric.GetTimer("path.request?path=" + pathEncoding)

		last := prev.Handler.(func(w http.ResponseWriter, r *http.Request, param RouterParam))
		return RouterMiddlewareContext{
			Data: prev.Data,
			Handler: func(w http.ResponseWriter, r *http.Request, param RouterParam) {
				begin := time.Now()
				last(w, r, param)
				end := time.Now()
				duration := end.Sub(begin)
				pathRequest.Update(duration)
			},
		}
	}
}
