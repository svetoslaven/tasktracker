package models

type TokenScope int8

const TokenScopeVerification TokenScope = 1

func (s TokenScope) String() string {
	switch s {
	case TokenScopeVerification:
		return "verification"
	default:
		panic("invalid token scope")
	}
}
