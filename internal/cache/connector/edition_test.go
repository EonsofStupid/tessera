package connector

import "testing"

func TestStartConnectorsForEdition(t *testing.T) {
	t.Parallel()
	conf := &CachesConfig{}
	conf.Connectors.Redis.Enabled = true
	if err := StartConnectorsForEdition("public", conf); err == nil {
		t.Fatal("public edition must refuse Redis")
	}
	if err := StartConnectorsForEdition("enterprise", conf); err != nil {
		t.Fatal(err)
	}
	conf.Connectors.Redis.Enabled = false
	if err := StartConnectorsForEdition("public", conf); err != nil {
		t.Fatal(err)
	}
}
