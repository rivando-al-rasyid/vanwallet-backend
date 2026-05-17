package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
)

type RootController struct{}

func NewRootController() *RootController {
	return &RootController{}
}

func (r *RootController) HelloKoda(c *gin.Context) {
	// mengirim response
	c.JSON(http.StatusOK, dto.Response{
		Message: "hello",
		Data: gin.H{
			"name": "koda",
		},
		Success: true,
		Error:   "",
	})
	// pastikan setelah response tidak ada proses logika
}

func (r *RootController) HelloString(c *gin.Context) {
	c.String(http.StatusOK, "%s", "hello")
}
