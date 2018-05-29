package auth

import (
	"errors"
	"gbird"
	"gbird/config"
	"gbird/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

//CustomClaims 自定义数据
type CustomClaims struct {
	Data interface{}
	jwt.StandardClaims
}

var (
	//ErrTokenExpired 过期
	ErrTokenExpired = errors.New("Token is expired")
	//ErrTokenNotValidYet yet
	ErrTokenNotValidYet = errors.New("Token not active yet")
	//ErrTokenMalformed fo
	ErrTokenMalformed = errors.New("That's not even a token")
	//ErrTokenInvalid invalid
	ErrTokenInvalid = errors.New("Couldn't handle this token")
	//SignKey 加密码
	SignKey = []byte(config.Config["jwtsecret"])
)

//CreateJWTToken 创建JWT Token
func CreateJWTToken(flag string, data interface{}) (string, error) {
	claims := CustomClaims{
		data,
		jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tstr, err := token.SignedString([]byte(SignKey))
	return flag + " " + tstr, err
}

func validation(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return SignKey, nil
		})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return token, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return token, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return token, ErrTokenNotValidYet
			}
			return token, ErrTokenInvalid
		}
	}
	return token, err
}
func getTokenData(token *jwt.Token) interface{} {
	if claims, ok := token.Claims.(*CustomClaims); ok {
		return claims.Data
	}
	return nil
}

func getToken(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		token = c.Request.Header.Get("token")
	}
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		token = c.Param("token")
	}
	if token == "" {
		token = c.PostForm("token")
	}
	return token
}

//ValidateTokenMiddleware 验证Token
func ValidateTokenMiddleware(needToken func(*gbird.Context) bool) func(*gbird.Context) {
	logger.Infoln("Token验证：开启")
	return gbird.GinToBird(
		func(c *gin.Context) {
			if c.Request.URL.Path != "/api/auth/login" && c.Request.URL.Path != "/api/auth/register" {
				if needToken == nil || needToken(&gbird.Context{Context: c}) {
					token := getToken(c)
					if token == "" {
						c.AbortWithStatusJSON(200, gin.H{"errcode": 1, "errmsg": "no token"})
						return
					}
					temps := strings.Split(token, " ")
					if len(temps) != 2 {
						c.AbortWithStatusJSON(200, gin.H{"errcode": 1, "errmsg": "no token"})
						return
					}
					_, err := validation(temps[1])
					if err != nil {
						c.AbortWithStatusJSON(200, gin.H{"errcode": 2, "errmsg": err.Error()})
						return
					}
				}
			}
			c.Next()
		})
}

//GetTokenData  获取token携带数据
func GetTokenData(c *gbird.Context) (string, interface{}) {
	token := getToken(c.Context)
	if token == "" {
		return "", nil
	}
	flag := strings.Split(token, " ")[0]
	tt, err := validation(strings.Split(token, " ")[1])
	if err != nil {
		return "", nil
	}
	return flag, getTokenData(tt)
}

func verify(c *gbird.Context) (interface{}, error) {
	token := getToken(c.Context)
	if token == "" {
		return nil, errors.New("no token")
	}
	tt, err := validation(strings.Split(token, " ")[1])
	if err != nil {
		return nil, err
	}
	data := getTokenData(tt)
	return data, nil
}

func refresh(tokenString string) (string, error) {
	tokens := strings.Split(tokenString, " ")
	flag := tokens[0]
	tt, err := validation(tokens[1])
	if err != nil {
		return "", err
	}
	data := getTokenData(tt)
	return CreateJWTToken(flag, data)
}
