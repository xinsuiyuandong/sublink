package routers

import (
	"sublink/api"
	"sublink/middlewares"
	"sublink/models" // 【新增】: 导入 models 包

	"github.com/gin-gonic/gin"
)

func Clients(r *gin.Engine) {
	ClientsGroup := r.Group("/c")
	ClientsGroup.Use(middlewares.GetIp)
	{
		// ClientsGroup.GET("/v2ray/:subname", api.GetV2ray)
		// ClientsGroup.GET("/clash/:subname", api.GetClash)
		// ClientsGroup.GET("/surge/:subname", api.GetSurge)
		ClientsGroup.GET("/", api.GetClient)
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
