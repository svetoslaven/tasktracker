package models

type TokenScope int8

const (
	TokenScopeVerification  TokenScope = 1
	TokenScopePasswordReset TokenScope = 2
)

func (s TokenScope) String() string {
	switch s {
	case TokenScopeVerification:
		return "verification"
	case TokenScopePasswordReset:
		return "password reset"
	default:
		panic("invalid token scope")
	}
}
