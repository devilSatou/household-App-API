package middleware

// TokenGeneratorInterface はトークン生成のインターフェース
type TokenGeneratorInterface interface {
	Generate(userID uint, email string) (string, error)
}

// TokenValidatorInterface はトークン検証のインターフェース
type TokenValidatorInterface interface {
	Validate(tokenString string) (*Claims, error)
}

// TokenGeneratorImpl は TokenGeneratorInterface の実装
type TokenGeneratorImpl struct{}

// NewTokenGenerator は TokenGeneratorInterface を返す
func NewTokenGenerator() TokenGeneratorInterface {
	return &TokenGeneratorImpl{}
}

func (g *TokenGeneratorImpl) Generate(userID uint, email string) (string, error) {
	return GenerateToken(userID, email)
}

// TokenValidatorImpl は TokenValidatorInterface の実装
type TokenValidatorImpl struct{}

// NewTokenValidator は TokenValidatorInterface を返す
func NewTokenValidator() TokenValidatorInterface {
	return &TokenValidatorImpl{}
}

func (v *TokenValidatorImpl) Validate(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString)
}
