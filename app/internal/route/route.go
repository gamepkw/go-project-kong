package route

import (
	"main/app/internal/handler"

	"github.com/labstack/echo/v4"
)

type Route struct {
	router  *echo.Echo
	handler handler.Handler
}

func New(router *echo.Echo, handler handler.Handler) Route {
	return Route{
		router:  router,
		handler: handler,
	}
}

func (r *Route) RegisterRoute() {

	apiGroup := r.router.Group("/api")
	apiGroup.POST("/test", r.handler.Test)
}
