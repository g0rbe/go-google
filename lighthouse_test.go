package google_test

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/g0rbe/go-google"
)

func TestRunLighthouse(t *testing.T) {

	res, err := google.RunLighthouse("https://gorbe.io/", nil, google.LighthouseCategoryAll...)
	if err != nil {
		t.Fatalf("FAIL: %s\n", err)
	}

	if res.RequestedURL().String() != "https://gorbe.io/" {
		t.Fatalf("Invalid RequestedURL: %s\n", res.RequestedURL().String())
	}

	t.Logf("performance='%d' accessibility='%d'  best-practices='%d' seo='%d' average='%d' total='%d'\n",
		res.Score("performance"), res.Score("accessibility"), res.Score("best-practices"), res.Score("seo"), res.Score("average"), res.Score("total"))
}

// func ExampleRunLighthouse() {

// 	r := google.RunLighthouse("https://gorbe.io/", nil, google.LighthouseCategoryAll...)
// 	if r.Errs() != nil {
// 		// Handle error
// 	}

// 	for k := range r.Categories {
// 		fmt.Printf("%s %d\n", k, int(r.Categories[k].Score*100))
// 	}
// }

func TestRunLighthouseErrLighthouseFailedDocumentRequest(t *testing.T) {

	_, err := google.RunLighthouse("https://gorbe.ioo", nil, google.LighthouseCategoryAll...)
	if err == nil {
		t.Fatalf("FAIL: error is nil\n")
	}

	if !errors.Is(err, google.ErrLighthouseFailedDocumentRequest) {
		if errors.Is(err, google.ErrLighthouseRateLimitExceeded) {
			t.Logf("ErrLighthouseRateLimitExceeded\n")
		} else {
			t.Fatalf("Unknown error: \"%s\"\n", err)
		}
	}
}

func TestRunLighthouseErrLighthouseInvalidKey(t *testing.T) {

	_, err := google.RunLighthouse("https://gorbe.io", google.NewApiKey("invalid"), google.LighthouseCategoryAll...)
	if err == nil {
		t.Fatalf("FAIL: error is nil\n")
	}

	if !errors.Is(err, google.ErrLighthouseInvalidKey) {
		if errors.Is(err, google.ErrLighthouseRateLimitExceeded) {
			t.Logf("ErrLighthouseRateLimitExceeded\n")
		} else {
			t.Fatalf("Unknown error: \"%s\"\n", err)
		}
	}
}

func TestRunLighthouseErrLighthouseInvalidCategory(t *testing.T) {

	_, err := google.RunLighthouse("https://gorbe.io", nil, google.LighthouseCategory("invalid"))
	if err == nil {
		t.Fatalf("FAIL: error is nil\n")
	}

	if !errors.Is(err, google.ErrLighthouseInvalidCategory) {
		if errors.Is(err, google.ErrLighthouseRateLimitExceeded) {
			t.Logf("ErrLighthouseRateLimitExceeded\n")
		} else {
			t.Fatalf("Unknown error: \"%s\"\n", err)
		}
	}
}

func TestRunLighthouseErrLighthouseInvalidStrategy(t *testing.T) {

	_, err := google.RunLighthouse("https://gorbe.io", nil, google.LighthouseStrategy("invalid"))
	if err == nil {
		t.Fatalf("FAIL: error is nil\n")
	}

	if !errors.Is(err, google.ErrLighthouseInvalidStrategy) {
		if errors.Is(err, google.ErrLighthouseRateLimitExceeded) {
			t.Logf("ErrLighthouseRateLimitExceeded\n")
		} else {
			t.Fatalf("Unknown error: \"%s\"\n", err)
		}
	}
}

func TestRunLighthouseErrLighthouseInvalidUrl(t *testing.T) {

	_, err := google.RunLighthouse("gorbe.io", nil, google.LighthouseCategoryAll...)
	if err == nil {
		t.Fatalf("FAIL: error is nil\n")
	}

	if !errors.Is(err, google.ErrLighthouseInvalidUrl) {

		if errors.Is(err, google.ErrLighthouseRateLimitExceeded) {
			t.Logf("ErrLighthouseRateLimitExceeded\n")
		} else {
			t.Fatalf("Unknown error: \"%s\"\n", err)
		}
	}
}

func TestRunConcurrentLighthouse(t *testing.T) {

	if runtime.NumCPU() < 16 {
		t.Skipf("Not enough number of CPU")
	}

	testurls := make([]string, 0, runtime.NumCPU()*2)

	for i := 0; i < runtime.NumCPU()*2; i++ {
		testurls = append(testurls, "https://example.com")
	}

	res, errs := google.RunConcurrentLighthouse(context.TODO(), runtime.NumCPU(), testurls, nil, google.LighthouseCategoryAll...)

	if len(res)+len(errs) != len(testurls) {
		t.Fatalf("FAIL: Invlalid result length: %d / %d\n", len(res)+len(errs), len(testurls))
	}

	for i := range errs {
		if !errors.Is(errs[i], google.ErrLighthouseUnprocessable) &&
			!errors.Is(errs[i], google.ErrLighthouseRateLimitExceeded) {

			t.Errorf("ERROR: %s\n", errs[i])
		}
	}
}
