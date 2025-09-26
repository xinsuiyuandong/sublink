package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// === Part 1: 用于二维码短链接生成 ===

// 〔中文注释〕: 定义接收 JSON 数据的结构体
type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

// 〔中文注释〕: /api/short 接口的处理函数
func GenerateShortLink(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效，需要一个 'url' 字段。"})
		return
	}

	// 〔中文注释〕: 调用 utils 包中现有的 Short 方法来生成短链接
	shortURL, err := utils.Short(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成短链接失败"})
		return
	}

	// 〔中文注释〕: 构造完整的短链接返回给 x-panel
	// 格式为: http://<你的sublink域名>:8000/s/短代码
	fullShortURL := "http://" + c.Request.Host + "/s/" + url.PathEscape(shortURL)

	// 〔中文注释〕: 以纯文本形式返回完整的短链接
	c.String(http.StatusOK, fullShortURL)
}


// === Part 2: 用于通用订阅转换 ===

// 〔中文注释〕: 定义接收订阅转换请求的结构体
type ConvertRequest struct {
	URL    string `json:"url" binding:"required"`
	Target string `json:"target" binding:"required"` // 'clash' or 'surge' etc.
}

// 〔中文注释〕: /api/convert 接口的处理函数
func ConvertSubscription(c *gin.Context) {
	var req ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效，需要 'url' 和 'target' 字段。"})
		return
	}

	// 1. 从用户提供的 URL 获取订阅内容
	resp, err := http.Get(req.URL)
	if err != nil {
		log.Printf("获取订阅链接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取原始订阅链接失败"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	
	// 解码内容（通常是 Base64）
	decodedBody := node.Base64Decode(string(body))
	// 按行分割成节点链接数组
	urls := strings.Split(decodedBody, "\n")
	
	// 2. 准备一个默认的转换配置 (如果需要更复杂的配置，可以从数据库读取)
	// 〔中文注释〕: 这里我们使用一个硬编码的默认配置，您也可以修改为从数据库模板读取
	defaultConfig := node.SqlConfig{
		Proxies: []string{}, // Proxies 会被自动填充
		ProxyGroup: []node.ProxyGroup{
			{Name: "PROXY", Type: "select", Proxies: []string{"自动选择", "DIRECT"}},
		},
		Rule: []string{"FINAL,PROXY"},
	}

	var result string
	var convertErr error

	// 3. 根据 target 调用不同的转换函数
	switch strings.ToLower(req.Target) {
	case "clash":
		// 复用 node.EncodeClash
		clashConfigBytes, err := node.EncodeClash(urls, defaultConfig)
		if err != nil {
			convertErr = err
		} else {
			result = string(clashConfigBytes)
		}
	case "surge":
		// 复用 node.EncodeSurge
		surgeConfig, err := node.EncodeSurge(urls, defaultConfig)
		if err != nil {
			convertErr = err
		} else {
			// 为 Surge 添加必要的头部信息
			header := "#!MANAGED-CONFIG " + "http://" + c.Request.Host + c.Request.RequestURI + " interval=86400 strict=false\n"
			result = header + surgeConfig
		}
	case "v2ray":
		// 对于 v2ray，通常是直接返回 base64 编码的节点列表
		result = node.Base64Encode(strings.Join(urls, "\n"))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的目标类型: " + req.Target})
		return
	}

	if convertErr != nil {
		log.Printf("转换订阅失败: %v", convertErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "转换订阅失败: " + convertErr.Error()})
		return
	}
	
	// 4. 将转换后的结果作为纯文本返回
	c.String(http.StatusOK, result)
}