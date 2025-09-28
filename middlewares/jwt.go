package middlewares

import (
	"errors"
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
// 【关键修改】：该函数已清空，现在它无条件地放行所有请求，彻底移除 JWT 验证。
func AuthorToken(c *gin.Context) {
	c.Next() // 立即调用 c.Next()，不进行任何检查
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
