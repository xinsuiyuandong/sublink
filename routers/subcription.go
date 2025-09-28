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

	// --- 【核心新增】: 处理短链接访问的公开路由 ---
	// 〔中文注释〕: 当手机 App 访问例如 /c/FjK8sLp3 时，这个路由会捕获请求
	r.GET("/c/:code", func(c *gin.Context) {
		// 1. 从 URL 路径中获取随机码
		code := c.Param("code")
		var sub models.Subcription

		// 2. 根据这个随机码去数据库的 'code' 字段中查找对应的记录
		if err := models.DB.Where("code = ?", code).First(&sub).Error; err != nil {
			log.Printf("短链接解析失败: 未找到代码 '%s', 错误: %v", code, err)
			c.String(http.StatusNotFound, "Link Not Found") // 〔中文注释〕: 如果找不到，返回 404
			return
		}

		// 3. 【重要】: 如果找到了记录，就将存储在 Config 字段中的原始长链接 (vless://...) 作为响应返回
		c.String(http.StatusOK, sub.Config)
	})
}
