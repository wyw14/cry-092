package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wyw14/cry-092/internal/application"
	"github.com/wyw14/cry-092/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
)

type Tokens struct {
	AccessToken, RefreshToken         string
	AccessExpiresAt, RefreshExpiresAt time.Time
}
type Service struct {
	Users      application.UserRepository
	Clock      application.Clock
	IDs        application.IDGenerator
	SigningKey []byte
	Issuer     string
}

func (s Service) Login(ctx context.Context, login, password string) (Tokens, error) {
	user, err := s.Users.FindByLogin(ctx, login)
	if err != nil {
		return Tokens{}, err
	}
	if !user.Active || bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)) != nil {
		return Tokens{}, fmt.Errorf("invalid credentials")
	}
	now := s.Clock.Now()
	accessExpiry, refreshExpiry := now.Add(15*time.Minute), now.Add(7*24*time.Hour)
	claims := jwt.MapClaims{"sub": user.ID, "iss": s.Issuer, "exp": accessExpiry.Unix(), "iat": now.Unix(), "roles": roleNames(user.Roles)}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	access, err := token.SignedString(s.SigningKey)
	if err != nil {
		return Tokens{}, err
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return Tokens{}, err
	}
	refresh := hex.EncodeToString(raw)
	digest := sha256.Sum256([]byte(refresh))
	if err := s.Users.SaveRefreshToken(ctx, identity.RefreshToken{ID: s.IDs.NewID(), UserID: user.ID, Digest: digest, ExpiresAt: refreshExpiry, Version: 1}); err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: refresh, AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}

func roleNames(roles map[identity.Role]bool) []string {
	names := make([]string, 0, len(roles))
	for role, enabled := range roles {
		if enabled {
			names = append(names, string(role))
		}
	}
	return names
}
