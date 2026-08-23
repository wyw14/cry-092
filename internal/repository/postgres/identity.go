package postgres

import (
	"context"

	"github.com/wyw14/cry-092/internal/domain/identity"
)

func (s *Store) FindByLogin(ctx context.Context, login string) (*identity.User, error) {
	var u identity.User
	var roles []string
	err := s.executor(ctx).QueryRow(ctx, `SELECT id,display_name,login,password_hash,roles,unit_id,active,version,created_at FROM users WHERE login=lower($1)`, login).Scan(&u.ID, &u.DisplayName, &u.Login, &u.PasswordHash, &roles, &u.UnitID, &u.Active, &u.Version, &u.CreatedAt)
	if err != nil {
		return nil, mapNotFound(err)
	}
	u.Roles = map[identity.Role]bool{}
	for _, role := range roles {
		u.Roles[identity.Role(role)] = true
	}
	return &u, nil
}
func (s *Store) SaveRefreshToken(ctx context.Context, t identity.RefreshToken) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO refresh_tokens(id,user_id,digest,expires_at,revoked_at,version) VALUES($1,$2,$3,$4,$5,$6)`, t.ID, t.UserID, t.Digest[:], t.ExpiresAt, t.RevokedAt, t.Version)
	return err
}
func (s *Store) FindRefreshToken(ctx context.Context, digest [32]byte) (*identity.RefreshToken, error) {
	var t identity.RefreshToken
	var raw []byte
	err := s.executor(ctx).QueryRow(ctx, `SELECT id,user_id,digest,expires_at,revoked_at,version FROM refresh_tokens WHERE digest=$1`, digest[:]).Scan(&t.ID, &t.UserID, &raw, &t.ExpiresAt, &t.RevokedAt, &t.Version)
	if err != nil {
		return nil, mapNotFound(err)
	}
	copy(t.Digest[:], raw)
	return &t, nil
}
func (s *Store) RevokeRefreshToken(ctx context.Context, id string, version int64) error {
	tag, err := s.executor(ctx).Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now(),version=version+1 WHERE id=$1 AND version=$2 AND revoked_at IS NULL`, id, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return mapNotFound(err)
	}
	return nil
}

func (s *Store) Enqueue(ctx context.Context, id, topic string, payload []byte) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO outbox_messages(id,topic,payload,attempts,available_at,created_at) VALUES($1,$2,$3,0,now(),now()) ON CONFLICT(id) DO NOTHING`, id, topic, payload)
	return err
}
