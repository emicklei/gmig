package main

import "testing"

func TestCmdTemplate(t *testing.T) {
	if err := newApp().Run([]string{"gmig", "template", "./test/template_test.txt"}); err != nil {
		t.Fatal("unexpected error", err)
	}
}
