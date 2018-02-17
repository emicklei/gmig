package gmig

import "testing"

func TestLoadStateFromGCS(t *testing.T) {
	gcs := GCS{Configuration: Config{Bucket: "gs://kramphub-gmig-toolshed-shared", StateObject: "gmig-last-migration"}}
	t.Log("save state")
	if err := gcs.SaveState("temp"); err != nil {
		t.Fatal(err)
	}
	t.Log("load state")
	v, err := gcs.LoadState()
	t.Log(v)
	t.Log(err)
}
