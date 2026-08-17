package core

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var usedMFAChallenges sync.Map

func claimUserID(value interface{}) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	default:
		var userId int
		fmt.Sscanf(fmt.Sprintf("%v", typed), "%d", &userId)
		return userId
	}
}

func (r *Runtime) generateJWT(userId int, email string, userName string, roleName string, isRefresh bool) interface{} {
	expirationTime := time.Now().Add(24 * 30 * time.Hour)
	tokenType := "access"
	if isRefresh {
		expirationTime = time.Now().Add(24 * 180 * time.Hour)
		tokenType = "refresh"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "joss_default_secret_change_in_production"
	}

	claims := jwt.MapClaims{
		"user_id":    userId,
		"email":      email,
		"name":       userName,
		"role":       roleName,
		"token_type": tokenType,
		"exp":        expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Printf("[Security] Error generando JWT: %v\n", err)
		return false
	}

	return tokenString
}

func (r *Runtime) generateMFAChallengeJWT(userId int, email string) interface{} {
	usedMFAChallenges.Range(func(key, value interface{}) bool {
		if usedAt, ok := value.(time.Time); ok && time.Since(usedAt) > 10*time.Minute {
			usedMFAChallenges.Delete(key)
		}
		return true
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "joss_default_secret_change_in_production"
	}

	claims := jwt.MapClaims{
		"user_id":    userId,
		"email":      email,
		"name":       "MFA_Pending",
		"role":       "guest",
		"token_type": "mfa_challenge",
		"jti":        uuid.New().String(),
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(5 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		LogError("[Security] Error generating MFA challenge: %v", err)
		return false
	}
	return tokenString
}

func (r *Runtime) parseJWT(tokenString string) (map[string]interface{}, bool) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "joss_default_secret_change_in_production"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		fmt.Printf("[ValidateJWT Error] Token Length: %d | Error: %v\n", len(tokenString), err)
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	}

	return nil, false
}

func (r *Runtime) ValidateJWT(tokenString string) (map[string]interface{}, bool) {
	claims, valid := r.parseJWT(tokenString)
	if !valid {
		return nil, false
	}
	if fmt.Sprintf("%v", claims["token_type"]) == "mfa_challenge" {
		return nil, false
	}
	return claims, true
}
