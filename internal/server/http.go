package server

import (
	"demo/api/v1"
	"demo/internal/conf"
	"demo/internal/service"
	"github.com/go-kratos/kratos/v2/errors"
	shttp "net/http"
	"reflect"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	opts = append(opts, http.ResponseEncoder(responseEncoder), http.ErrorEncoder(errorEncoder))
	srv := http.NewServer(opts...)
	v1.RegisterGreeterSvcHTTPServer(srv, greeter)
	return srv
}

type BizResp interface {
	GetRet() *v1.BaseResp
}

func responseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(http.Redirector); ok {
		url, code := rd.Redirect()
		shttp.Redirect(w, r, url, code)
		return nil
	}
	bizResp, ok := v.(BizResp)
	if ok && bizResp != nil && bizResp.GetRet() == nil {
		val := reflect.ValueOf(v).Elem()
		ret := val.FieldByName("Ret")
		if ret.CanSet() {
			ret.Set(reflect.ValueOf(&v1.BaseResp{Code: 0, Msg: "", Reason: ""}))
		}
	}
	codec, _ := http.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

func errorEncoder(w shttp.ResponseWriter, r *shttp.Request, err error) {
	se := errors.FromError(err)
	codec, _ := http.CodecForRequest(r, "Accept")
	bizCode := v1.ErrorReason_value[se.Reason]
	if bizCode == 0 {
		bizCode = se.GetCode()
	}
	rsp := v1.Response{
		Ret: &v1.BaseResp{
			Code:   bizCode,
			Reason: se.Reason,
			Msg:    se.Message,
		},
	}
	body, err := codec.Marshal(&rsp)
	if err != nil {
		w.WriteHeader(shttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(se.Code))
	_, _ = w.Write(body)
}
