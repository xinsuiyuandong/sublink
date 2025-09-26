package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

func Total(r gin.IRoutes) {
	TotalGroup := r.privateGroup("/api/v1/total")
	{
		TotalGroup.GET("/sub", api.SubTotal)
		TotalGroup.GET("/node", api.NodesTotal)
	}

}
