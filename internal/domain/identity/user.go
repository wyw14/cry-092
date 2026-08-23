package identity

import (
	"fmt"
	"strings"
	"time"
)

type Role string

const (
	RoleRepresentative Role = "representative"
	RoleUnitOfficer    Role = "unit_officer"
	RoleSupervisor     Role = "supervisor"
	RoleAdministrator  Role = "administrator"
)

type User struct {
	ID           string
	DisplayName  string
	Login        string
	PasswordHash []byte
	Roles        map[Role]bool
	UnitID       string
	Active       bool
	Version      int64
	CreatedAt    time.Time
}

func NewUser(id, name, login string, passwordHash []byte, roles []Role, unitID string, now time.Time) (*User, error) {
	if id == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(login) == "" || len(passwordHash) == 0 {
		return nil, fmt.Errorf("user identity fields are required")
	}
	roleSet := make(map[Role]bool, len(roles))
	for _, role := range roles {
		if !validRole(role) {
			return nil, fmt.Errorf("unknown role %q", role)
		}
		roleSet[role] = true
	}
	if len(roleSet) == 0 {
		return nil, fmt.Errorf("at least one role is required")
	}
	return &User{ID: id, DisplayName: strings.TrimSpace(name), Login: strings.ToLower(strings.TrimSpace(login)), PasswordHash: append([]byte(nil), passwordHash...), Roles: roleSet, UnitID: unitID, Active: true, Version: 1, CreatedAt: now.UTC()}, nil
}

func validRole(role Role) bool {
	switch role {
	case RoleRepresentative, RoleUnitOfficer, RoleSupervisor, RoleAdministrator:
		return true
	default:
		return false
	}
}

func (u User) HasRole(role Role) bool { return u.Active && u.Roles[role] }

func (u User) CanManageUnit(unitID string) bool {
	return u.Active && u.UnitID != "" && u.UnitID == unitID && u.Roles[RoleUnitOfficer]
}
