package main

import (
    "crypto/rand"
    "sync"
    "encoding/base64"
    "time"
    "github.com/dgrijalva/jwt-go"
    "fmt"
    "errors"
)

var once    sync.Once
type authSecret struct {
    Value   string
}

func NewTokenSecret() *authSecret {
    key := make([]byte, 64)
    _, err := rand.Read(key)
    if err != nil {
        Logs.Errorf("Unable to create secret key for application! What?")
        panic(err)
    }
    secret := &authSecret{}
    once.Do(func() {
        secret.Value = base64.StdEncoding.EncodeToString([]byte(key))
    })
    return secret
}

func (a *authSecret) MakeToken(userName string, rbac string, lenMinutes ...int) (interface{}, error) {

    minutes := 60
    if len(lenMinutes) > 0 {
        minutes = lenMinutes[0]
    }
    token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), jwt.MapClaims{
        "username": userName,
        "rbac": rbac,
        "exp": time.Now().Add(time.Minute * time.Duration(minutes)),
    })
    tokenString, err := token.SignedString([]byte(a.Value))
    if err != nil {
        Logs.Errorf("Unable to create application authentication token!")
        return "", err
    }
    data := map[string]string {
        "token": tokenString,
    }
    return data, nil
}

func (a *authSecret) ValidateToken(tokenString string) (interface{}, error) {
    token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected signing method: %+v", token.Header["alg"])
        }
        return []byte(a.Value), nil
    })
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, errors.New("Expired token")
    } else {
        expiry, _ := time.Parse(time.RFC3339, claims["exp"].(string))
        if expiry.Before(time.Now()) {
            return nil, errors.New("Expired token")
        }
        return token, nil
    }
}

func (a *authSecret) GetClaims(tokenString string) (jwt.MapClaims) {
    token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected signing method: %+v", token.Header["alg"])
        }
        return []byte(a.Value), nil
    })
    claims, _ := token.Claims.(jwt.MapClaims)
    return claims
}
