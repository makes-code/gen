// This file is generated by makes-code ... do not edit

package types

import (
	"github.com/makes-code/gen/test/user"
)

type Users []User

type userData struct {
	id         string
	name       string
	identities []user.Identity
	profile    user.Profile
	workspaces map[string]user.Workspace
}

func (u *userData) ID() string                            { return u.id }
func (u *userData) Name() string                          { return u.name }
func (u *userData) Identities() []user.Identity           { return u.identities }
func (u *userData) Profile() user.Profile                 { return u.profile }
func (u *userData) Workspaces() map[string]user.Workspace { return u.workspaces }
func (u *userData) Builder() *UserBuilder {
	return NewUserBuilder().
		WithID(u.id).
		WithName(u.name).
		WithIdentities(u.identities).
		WithProfile(u.profile).
		WithWorkspaces(u.workspaces)
}

// UserBuilder is a user builder
type UserBuilder struct {
	data userData
}

// NewUserBuilder returns a new user builder
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{}
}

// WithID sets the user id
func (builder *UserBuilder) WithID(id string) *UserBuilder {
	builder.data.id = id
	return builder
}

// WithName sets the user name
func (builder *UserBuilder) WithName(name string) *UserBuilder {
	builder.data.name = name
	return builder
}

// WithIdentities sets the user identities
func (builder *UserBuilder) WithIdentities(identities []user.Identity) *UserBuilder {
	builder.data.identities = identities
	return builder
}

// WithProfile sets the user profile
func (builder *UserBuilder) WithProfile(profile user.Profile) *UserBuilder {
	builder.data.profile = profile
	return builder
}

// WithWorkspaces sets the user workspaces
func (builder *UserBuilder) WithWorkspaces(workspaces map[string]user.Workspace) *UserBuilder {
	builder.data.workspaces = workspaces
	return builder
}

// Data returns the user data
func (builder *UserBuilder) Data() User { return &builder.data }

// Build validates and returns the built user
func (builder *UserBuilder) Build() (User, error) {
	if err := prebuild(builder); err != nil {
		return nil, err
	}
	return &builder.data, nil
}

// MustBuild returns the built user and panics if any validation error occurs
func (builder *UserBuilder) MustBuild() User {
	built, err := builder.Build()
	if err != nil {
		panic("failed to build user: " + err.Error())
	}
	return built
}