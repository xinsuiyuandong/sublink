package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"sublink/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// 【新增核心常量】: 固定的公共 API 密钥。前端和后端必须使用此密钥。
const StaticPublicAPIKey = "Your_Fixed_Static_API_Key_For_Sublink_is_20251010_xyz"
// 【新增常量】: Authorization 头部的前缀格式
const BearerPrefix = "Bearer "

// 随机密钥
// var Secret = []byte("sublink") // 秘钥
var Secret = []byte(models.ReadConfig().JwtSecret) // 从配置文件读取JWT密钥

// JwtClaims jwt声明
type JwtClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// AuthorToken 验证token中间件
func AuthorToken(c *gin.Context) {
    // 1. OPTIONS 预检直接通过
    if c.Request.Method == http.MethodOptions {
        c.Next()
        return
    }

    // 获取请求路径和 Authorization 头部
    path := c.Request.URL.Path
    authHeader := c.GetHeader("Authorization") // 获取 Authorization 头部，用于 A 和 C 逻辑

    // ----------------------------------------------------
    // 【优先级 A】：固定密钥验证 (针对 /api/short 和 /api/convert)
    // ----------------------------------------------------
    if path == "/api/short" || path == "/api/convert" {
        // 检查是否以 Bearer 开头
        if strings.HasPrefix(authHeader, BearerPrefix) {
            // 提取密钥
            submittedKey := strings.TrimSpace(strings.Replace(authHeader, BearerPrefix, "", 1))
            
            // 比对密钥是否匹配我们写死的常量
            if submittedKey == StaticPublicAPIKey { 
                c.Next() // 验证通过，立即放行
                return
            }
        }
        
        // 验证失败或格式错误，返回未授权错误，不继续执行下面的任何逻辑
        c.JSON(http.StatusUnauthorized, gin.H{"msg": "公共 API 密钥验证失败，请联系管理员"})
        c.Abort()
        return
    }

    // ----------------------------------------------------
    // 【优先级 B】：完全公开白名单 (不需要任何验证)
    // ----------------------------------------------------
    // 注意：/api/short 和 /api/convert 已在上面处理，不需要出现在这里。
    list := []string{
        "/static", 
        "/api/v1/auth/login", 
        "/api/v1/auth/captcha", 
        "/c/",
    }

    // 如果首页直接跳过
    if path == "/" {
        c.Next()
        return
    }

    // 如果是白名单直接跳过
    for _, v := range list {
        if strings.HasPrefix(path, v) {
            c.Next()
            return
        }
    }

    // ----------------------------------------------------
    // 【优先级 C】：JWT Token 检查（如果请求到达这里，说明它必须是私有路由）
    // ----------------------------------------------------
    
    // 如果 Authorization 头部为空，返回 401 Unauthorized
    if authHeader == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"msg": "请求未携带token"})
        c.Abort()
        return
    }

    // 去掉 Bearer 前缀并 trim 空格
    token := strings.TrimSpace(strings.Replace(authHeader, BearerPrefix, "", 1))

    // 检查 token 格式：必须包含 3 个点分隔的部分
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        // 使用标准的 401 状态码，并保持原有的错误信息
        c.JSON(http.StatusUnauthorized, gin.H{"msg": "token格式错误"}) 
        c.Abort()
        return
    }

    // 解析并校验 token
    mc, err := ParseToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "code": http.StatusUnauthorized, // 使用 http.StatusUnauthorized (401)
            "msg":  err.Error(),
        })
        c.Abort()
        return
    }

    c.Set("username", mc.Username)
    c.Next()
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*JwtClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
