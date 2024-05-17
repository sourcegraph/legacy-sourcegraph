package subscriptionsservice

import (
	"net/http"

	"github.com/sourcegraph/log"

	subscriptionsv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1/v1connect"
)

func RegisterV1(logger log.Logger, mux *http.ServeMux) {
	mux.Handle(subscriptionsv1connect.NewSubscriptionsServiceHandler(newHandlerV1(logger)))
}

type handlerV1 struct {
	subscriptionsv1connect.UnimplementedSubscriptionsServiceHandler

	logger log.Logger
}

var _ subscriptionsv1connect.SubscriptionsServiceHandler = (*handlerV1)(nil)

func newHandlerV1(logger log.Logger) *handlerV1 {
	return &handlerV1{
		logger: logger.Scoped("subscriptions.v1"),
	}
}
