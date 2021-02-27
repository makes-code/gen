package types

import "github.com/makes-code/gen/test/user"

type User interface {
	ID() string
	Name() string
	Identities() []user.Identity
	Profile() user.Profile
	Workspaces() map[string]user.Workspace
}

func (builder *UserBuilder) Prebuild() error {
	return nil
}

//go:generate go run ../main.go type model -repo test -name User
//go:generate go run ../main.go type payload -repo test -name User -tag Partial -strict -i ID -i Name=n
//go:generate go run ../main.go type document -repo test -name User -tag Partial -i Name=n -x Identities -x Profile -x Workspaces
