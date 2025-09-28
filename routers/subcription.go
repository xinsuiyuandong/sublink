package routers

import (
	"log"
	"net/http"
	"sublink/api"
	"sublink/models" // 【新增】: 导入 models 包

	"github.com/gin-gonic/gin"
)

func Subcription(r *gin.Engine) {
	// --- 您原有的后台管理 API 路由 (保持不变) ---
	SubcriptionGroup := r.Group("/api/v1/subcription")
	{
		SubcriptionGroup.POST("/add", api.SubAdd)
		SubcriptionGroup.DELETE("/delete", api.SubDel)
		SubcriptionGroup.GET("/get", api.SubGet)
		SubcriptionGroup.POST("/update", api.SubUpdate)
	}
}
