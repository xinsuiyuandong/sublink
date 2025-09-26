package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"sublink/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

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
    // 1) 允许 OPTIONS 预检直接通过（避免被 JWT 拦截）
    if c.Request.Method == http.MethodOptions {
        c.Next()
        return
    }

    // 2) 白名单：静态资源、登录、验证码，以及我们要公开的 /api/short 和 /api/convert
    list := []string{
        "/static",
        "/api/v1/auth/login",
        "/api/v1/auth/captcha",
        "/c/",
        "/api/v1/version",
        "/api/short",    // 允许生成短链的公开接口
        "/api/convert",  // 允许订阅转换的公开接口
    }

    // 如果首页直接跳过
    if c.Request.URL.Path == "/" {
        c.Next()
        return
    }

    // 如果是白名单直接跳过
    for _, v := range list {
        if strings.HasPrefix(c.Request.URL.Path, v) {
            c.Next()
            return
        }
    }

    // 3) 正常走 token 校验
    token := c.GetHeader("Authorization")
    if token == "" {
        c.JSON(400, gin.H{"msg": "请求未携带token"})
        c.Abort()
        return
    }

    // 去掉 Bearer 前缀并 trim 空格（先去掉前缀，再按 '.' 分段）
    token = strings.TrimSpace(strings.Replace(token, "Bearer ", "", 1))

    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        c.JSON(400, gin.H{"msg": "token格式错误"})
        c.Abort()
        return
    }

    mc, err := ParseToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "code": 401,
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
