package main

import "testing"

func Test_loadbalancerURLMap_patchPathsAndService(t *testing.T) {
	m := loadbalancerURLMap{
		PathMatchers: []pathMatcher{
			{
				Name: "a",
				PathRules: []pathsAndService{{
					Paths:   []string{"/a"},
					Service: "theA",
				}},
			},
		},
	}
	err := m.patchPathsAndService(true, "a", "theA", []string{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(m.PathMatchers[0].PathRules), 0; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	{
		err := m.patchPathsAndService(false, "a", "theB", []string{"/b"}, false)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(m.PathMatchers[0].PathRules), 1; got != want {
			t.Log(m)
			t.Errorf("got [%v] want [%v]", got, want)
		}
	}
}
