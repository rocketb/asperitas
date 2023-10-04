package user

import "errors"

// Set of possible roles for a user.
var (
	RoleAdmin = Role{"ADMIN"}
	RoleUser  = Role{"USER"}
)

// Known roles in the syste.
var roles = map[string]Role{
	RoleAdmin.name: RoleAdmin,
	RoleUser.name:  RoleUser,
}

// Role Represents a user role in the system.
type Role struct {
	name string
}

func (r Role) Name() string {
	return r.name

} // UnmarshalText implement the unmarshal interface for JSON conversions.
func (r *Role) UnmarshalText(data []byte) error {
	r.name = string(data)
	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.name), nil
}

// ParseRole get the role name and return it if exist.
func ParseRole(name string) (Role, error) {
	role, ok := roles[name]
	if !ok {
		return Role{}, errors.New("invalid role")
	}

	return role, nil
}
