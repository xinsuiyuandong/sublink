package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

func Total(r *gin.Engine) {
	TotalGroup := r.Group("/total")
	{
		TotalGroup.GET("/sub", api.SubTotal)
		TotalGroup.GET("/node", api.NodesTotal)
	}

}
