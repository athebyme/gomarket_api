package config

import "os"

type JwtConfig struct {
	JWTSecret          string
	AccessTokenExpiry  int // в минутах
	RefreshTokenExpiry int // в днях
	DBConnection       string
}

func Load() *JwtConfig {
	return &JwtConfig{
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AccessTokenExpiry:  15, // 15 минут
		RefreshTokenExpiry: 7,  // 7 дней
		DBConnection:       os.Getenv("DB_CONNECTION"),
	}
}
