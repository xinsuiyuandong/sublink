package api

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
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

	// 复用 utils.RandString 生成随机短码（使用 16 以获得合理长度随机字符串）
	shortURL := utils.RandString(16)

	// 添加简单检查（尽管 RandString 不会失败，但更为鲁棒）
	if shortURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成短链接失败"})
		return
	}
    
	// 【健壮性改进】：确保只取 Hostname，避免出现如 "test.wudust.top:8000:8000" 这种双重端口错误
	host := c.Request.Host
	// 尝试切割主机和端口
	if h, _, err := strings.Cut(host, ":"); err == false {
		// 如果 Host 字段不包含端口，则使用完整 Host
	} else {
		// 如果包含端口，只使用 hostname
		host = h
	}

	// 构造完整的短链接返回给 x-panel
	// 格式为: http://<你的sublink域名>:8000/s/短代码
	// 注意：这里的端口 8000 是硬编码的，如果您的服务不是 8000，需要修改。
	fullShortURL := "http://" + host + ":8000" + "/s/" + url.PathEscape(shortURL)

	// 以纯文本形式返回完整的短链接
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
