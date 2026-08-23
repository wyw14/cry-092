package identity

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	Digest    [32]byte
	ExpiresAt time.Time
	RevokedAt *time.Time
	Version   int64
}

func (t RefreshToken) Active(now time.Time) bool {
	return t.RevokedAt == nil && now.UTC().Before(t.ExpiresAt.UTC())
}

func (t *RefreshToken) Revoke(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	at := now.UTC()
	t.RevokedAt = &at
	t.Version++
	return true
}
