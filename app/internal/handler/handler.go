package handler

import (
	"main/app/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service service.Service
}

func New(service service.Service) Handler {
	return &handler{
		service: service,
	}
}

type Handler interface {
	Test(c echo.Context) error
}

func (h *handler) Test(c echo.Context) error {

	// ctx := c.Request().Context()

	// ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()

	// if err := h.service.ProduceMessage(ctxWithTimeout); err != nil {
	// 	return c.JSON(http.StatusInternalServerError, nil)
	// }
	return c.JSON(http.StatusOK, nil)
}
