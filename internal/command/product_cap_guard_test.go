package command

import (
	"context"
	"errors"
	"testing"
)

type staticCapGuard struct {
	user     error
	org      error
	instance error
}

func (g staticCapGuard) DenyNewUser(context.Context) error { return g.user }
func (g staticCapGuard) DenyNewOrganization(context.Context) error {
	return g.org
}
func (g staticCapGuard) DenyNewInstance(context.Context) error {
	return g.instance
}

func TestAddOrgHonorsProductCapGuard(t *testing.T) {
	t.Parallel()
	commands := &Commands{productCapGuard: staticCapGuard{org: errors.New("demo_cap_exceeded")}}
	_, err := commands.AddOrg(context.Background(), "Second Org", "user", "instance", nil)
	if err == nil || err.Error() != "demo_cap_exceeded" {
		t.Fatalf("err %v", err)
	}
}

func TestAddHumanHonorsProductCapGuard(t *testing.T) {
	t.Parallel()
	commands := &Commands{productCapGuard: staticCapGuard{user: errors.New("demo_cap_exceeded")}}
	err := commands.AddHuman(context.Background(), "org", &AddHuman{Username: "user"}, false)
	if err == nil || err.Error() != "demo_cap_exceeded" {
		t.Fatalf("err %v", err)
	}
}
