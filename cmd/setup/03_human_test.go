package setup

import (
	"testing"

	"github.com/shippinAI/nomen/internal/command"
)

func TestApplyFirstHumanFromDeployEnvOmitsHumanWithoutPassword(t *testing.T) {
	t.Setenv("NOMEN_FIRSTINSTANCE_ORG_HUMAN_PASSWORD", "")
	org := command.InstanceOrgSetup{
		Name: "Nomen",
		Human: &command.AddHuman{
			Username:  "nomen-admin",
			FirstName: "Nomen",
			LastName:  "Admin",
		},
	}
	if err := applyFirstHumanFromDeployEnv(&org); err != nil {
		t.Fatal(err)
	}
	if org.Human != nil {
		t.Fatal("human must be omitted when deploy-time password is unset")
	}
}

func TestApplyFirstHumanFromDeployEnvRefusesStepsPassword(t *testing.T) {
	org := command.InstanceOrgSetup{
		Human: &command.AddHuman{Password: "Password1!"},
	}
	if err := applyFirstHumanFromDeployEnv(&org); err == nil {
		t.Fatal("expected refusal of a password stored in setup steps")
	}
}

func TestApplyFirstHumanFromDeployEnvKeepsYamlIdentityWithEnvPassword(t *testing.T) {
	t.Setenv("NOMEN_FIRSTINSTANCE_ORG_HUMAN_PASSWORD", "deploy-only-secret")
	org := command.InstanceOrgSetup{
		Human: &command.AddHuman{
			Username:  "owner@example.test",
			FirstName: "First",
			LastName:  "Owner",
		},
	}
	if err := applyFirstHumanFromDeployEnv(&org); err != nil {
		t.Fatal(err)
	}
	if org.Human.Username != "owner@example.test" || org.Human.Password != "deploy-only-secret" {
		t.Fatalf("%+v", org.Human)
	}
}

func TestApplyFirstHumanFromDeployEnvUsesEnvPassword(t *testing.T) {
	t.Setenv("NOMEN_FIRSTINSTANCE_ORG_HUMAN_PASSWORD", "deploy-only-secret")
	t.Setenv("NOMEN_FIRSTINSTANCE_ORG_HUMAN_USERNAME", "owner@example.test")
	org := command.InstanceOrgSetup{}
	if err := applyFirstHumanFromDeployEnv(&org); err != nil {
		t.Fatal(err)
	}
	if org.Human == nil {
		t.Fatal("expected human from deploy-time env")
	}
	if org.Human.Password != "deploy-only-secret" {
		t.Fatalf("password %q", org.Human.Password)
	}
	if org.Human.Username != "owner@example.test" {
		t.Fatalf("username %q", org.Human.Username)
	}
}
