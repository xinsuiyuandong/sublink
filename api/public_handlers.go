package api

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sublink/node"
	"sublink/utils"
	"sublink/models" // 【新增】: 导入 models 包

	"github.com/gin-gonic/gin"
)

// === Part 1: 用于二维码短链接生成 ===

// 〔中文注释〕: 定义接收 JSON 数据的结构体
type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

// 〔中文注释〕: 【修复后】的 /api/short 接口处理函数
func GenerateShortLink(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效，需要一个 'url' 字段。"})
		return
	}

	// 1. 生成一个唯一的随机短码
	shortCode := utils.RandString(16)

	// 2. 【核心】: 创建一个新的 Subcription 对象来存储映射关系
	subEntry := models.Subcription{
		Name:   "OneClick-" + shortCode, // 〔中文注释〕: 给个名字，方便后台识别
		Code:   shortCode,               // 〔中文注释〕: 存储随机短码
		Config: req.URL,                 // 【重要】: 存储原始的长链接 (vless://...)
	}

	// 3. 【核心】: 将这个映射关系保存到数据库
	if err := models.DB.Create(&subEntry).Error; err != nil {
		log.Printf("保存短链接到数据库失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建短链接失败"})
		return
	}

	// 4. 〔中文注释〕: 拼接并返回可以正常工作的短链接
	host := c.Request.Host
	if h, _, err := strings.Cut(host, ":"); err == false {
	} else {
		host = h
	}

	// 〔中文注释〕: 这里的端口 8000 是硬编码的，请确保您的 sublink 服务运行在该端口
	fullShortURL := "http://" + host + ":8000" + "/c/" + url.PathEscape(shortCode)
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
		// 【改进】：日志记录更详细的连接错误
		log.Printf("获取订阅链接失败 (连接或网络错误): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取原始订阅链接失败 (请检查网络或URL)"})
		return
	}
	defer resp.Body.Close()
    
	// 【关键修复】：检查上游服务器返回的状态码，解决 500 错误
	if resp.StatusCode != http.StatusOK {
		log.Printf("获取订阅链接返回状态码错误: %s", resp.Status)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取原始订阅链接失败，上游服务器返回状态码: " + resp.Status})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取订阅内容失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取订阅内容失败"})
		return
	}

	// 解码内容（通常是 Base64），复用 node.Base64Decode
	decodedBody := node.Base64Decode(string(body))
	// 按行分割成节点链接数组
	urls := strings.Split(strings.TrimSpace(decodedBody), "\n")

	// 2. 准备一个默认的转换配置（复用 node.SqlConfig 现有字段；Clash 和 Surge 为模板路径）
	// 〔中文注释〕: 这里使用硬编码默认值，也可以修改为从数据库或 models 读取
	defaultConfig := node.SqlConfig{
		Clash: "./template/clash.yaml", // 默认 Clash 模板路径（基于 Templateinit() 初始化）
		Surge: "./template/surge.conf", // 默认 Surge 模板路径（基于 Templateinit() 初始化）
		Udp:   false,                  // 默认不启用 UDP
		Cert:  false,                  // 默认不跳过证书验证
	}

	var result string
	var convertErr error

	// 3. 根据 target 调用不同的转换函数（复用 node.EncodeClash 和 node.EncodeSurge）
	switch strings.ToLower(req.Target) {
	case "clash":
		// 复用 node.EncodeClash（使用 urls 和 defaultConfig；模板中处理 groups/rules）
		clashConfigBytes, err := node.EncodeClash(urls, defaultConfig)
		if err != nil {
			convertErr = err
		} else {
			result = string(clashConfigBytes)
		}
	case "surge":
		// 复用 node.EncodeSurge（使用 urls 和 defaultConfig；模板中处理 groups/rules）
		surgeConfig, err := node.EncodeSurge(urls, defaultConfig)
		if err != nil {
			convertErr = err
		} else {
			// 为 Surge 添加必要的头部信息
			header := "#!MANAGED-CONFIG " + "http://" + c.Request.Host + c.Request.RequestURI + " interval=86400 strict=false\n"
			result = header + surgeConfig
		}
	case "v2ray":
		// 对于 v2ray，通常是直接返回 base64 编码的节点列表，复用 node.Base64Encode
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


