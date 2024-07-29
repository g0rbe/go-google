package google_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/g0rbe/go-google"
)

func TestRunLighthouse(t *testing.T) {

	r := google.RunLighthouse("https://gorbe.io/", google.NewApiKey(os.Getenv("GOOGLE_CLOUD_KEY")), google.LighthouseCategoryAll...)
	if r.Errs() != nil {
		t.Fatalf("FAIL: %#v\n", r.Errs())
	}

	if r.RequestedUrl != "https://gorbe.io/" {
		t.Fatalf("Invalid RequestedURL: %s\n", r.RequestedUrl)
	}
}

func ExampleRunLighthouse() {

	r := google.RunLighthouse("https://gorbe.io/", nil, google.LighthouseCategoryAll...)
	if r.Errs() != nil {
		// Handle error
	}

	for k := range r.Categories {
		fmt.Printf("%s %d\n", k, int(r.Categories[k].Score*100))
	}
}

func TestRunLighthouseFail(t *testing.T) {

	r := google.RunLighthouse("https://gorbe.ioo", google.NewApiKey(os.Getenv("GOOGLE_CLOUD_KEY")), google.LighthouseCategoryAll...)
	if r.Errs() == nil {
		t.Fatalf("FAIL: error is nil\n")
	}

	if r.RequestedUrl != "https://gorbe.ioo" {
		t.Fatalf("Invalid RequestedURL: %s\n", r.RequestedUrl)
	}
}

func TestRunConcurrentLighthouse(t *testing.T) {

	listData, err := os.ReadFile("test.list")
	if err != nil {
		t.Fatalf("Failed to open \"test.list\": %s\n", err)
	}

	testList := strings.Split(strings.TrimSpace(string(listData)), "\n")

	r := google.RunConcurrentLighthouse(context.TODO(), 16, testList, google.NewApiKey(os.Getenv("GOOGLE_CLOUD_KEY")), google.LighthouseCategoryAll...)

	if len(r) != len(testList) {
		t.Fatalf("FAIL: Invlalid result length: %d / %d\n", len(r), len(testList))
	}
	for i := range r {
		if r[i].Errs() != nil {
			t.Errorf("FAIL:%s -> %#v\n", r[i].RequestedUrl, r[i].Errs())
			continue
		}

		if r[i].FinalUrl != r[i].RequestedUrl {
			t.Errorf("FAIL: %s != %s\n", r[i].FinalUrl, r[i].RequestedUrl)
		}

	}
}
