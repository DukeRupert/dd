package handlers

import (
	"net/http"

	"github.com/dukerupert/dd/views"
	
	"github.com/labstack/echo/v4"
	"github.com/dukerupert/dd/pb"
)

type APIHandler struct {
	apiClient *pocketbase.Client
}

func NewAPIHandler(apiClient *pocketbase.Client) *APIHandler {
	return &APIHandler{
		apiClient: apiClient,
	}
}

func LoginHandler(c echo.Context) error {
	return views.Login().Render(c.Request().Context(), c.Response().Writer)
}

func (h *APIHandler) GetResource(c echo.Context) error {
	// id := c.Param("id")
	
	// resource, err := h.apiClient.GetResource(c.Request().Context(), id)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{
	// 		"error": "Failed to fetch resource",
	// 	})
	// }
	
	return c.JSON(http.StatusOK, nil)
}