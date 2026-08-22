package auth

import (
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("spide-pos-secret-key-2024")

type Claims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Name     string `json:"name"`
    Role     string `json:"role"`
    ShopID   int    `json:"shop_id"`
    ShopName string `json:"shop_name"`
    jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int, username, name, role string, shopID int, shopName string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Name:     name,
        Role:     role,
        ShopID:   shopID,
        ShopName: shopName,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "spide-pos",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

