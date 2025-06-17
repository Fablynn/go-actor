package token

import (
	"poker_server/common/pb"
	"poker_server/library/uerror"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	secretKey = make([]byte, 32)
)

func Init(keys string) {
	secretKey = []byte(keys)
}

type Token struct {
	jwt.RegisteredClaims
	Uid    uint64 `json:"user_id"`
	RoomId uint64 `json:"room_id"`
}

func GenToken(tt *Token) (string, error) {
	tt.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    "poker_server",                                          // 签发者
		IssuedAt:  jwt.NewNumericDate(time.Now()),                          // 签发时间
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1500000 * time.Hour)), // 过期时间
	}

	// 2. 生成Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tt)

	// 3. 签名（使用环境变量获取密钥更安全）
	return token.SignedString(secretKey)
}

func cb(token *jwt.Token) (interface{}, error) {
	// 验证签名算法
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, uerror.New(1, pb.ErrorCode_TYPE_ASSERT_FAILED, "unexpected signing method: %v", token.Header["alg"])
	}
	return secretKey, nil
}

func ParseToken(tt string) (*Token, error) {
	tok, err := jwt.ParseWithClaims(tt, &Token{}, cb)
	if err != nil {
		return nil, err
	}
	if ret, ok := tok.Claims.(*Token); ok && tok.Valid {
		return ret, nil
	}
	return nil, uerror.New(1, pb.ErrorCode_PARSE_FAILED, "token is invalid")
}
