package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

// Possible values for category paramater
var (
	LighthouseCategoryAccessibility = NewLighthouseCategory("ACCESSIBILITY")
	LighthouseCategoryBestPractices = NewLighthouseCategory("BEST_PRACTICES")
	LighthouseCategoryPerformance   = NewLighthouseCategory("PERFORMANCE")
	LighthouseCategorySEO           = NewLighthouseCategory("SEO")

	// Shorthand to all PagespeedCategory
	LighthouseCategoryAll = []LighthouseParam{LighthouseCategoryAccessibility, LighthouseCategoryBestPractices, LighthouseCategoryPerformance, LighthouseCategorySEO}
)

// Possible values for strategy paramater
var (
	LighthouseStrategyDesktop = NewLighthouseStrategy("dektop")
	LighthouseStrategyMobile  = NewLighthouseStrategy("mobile")
)

// LighthouseParam stores a single request param as a key/value pair.
type LighthouseParam struct {
	k, v string
}

func NewLighthouseCategory(v string) LighthouseParam {
	return LighthouseParam{k: "category", v: v}
}

func NewLighthouseLocale(v string) LighthouseParam {
	return LighthouseParam{k: "locale", v: v}
}

func NewLighthouseStrategy(v string) LighthouseParam {
	return LighthouseParam{k: "strategy", v: v}
}

func NewLighthouseUTMCampaign(v string) LighthouseParam {
	return LighthouseParam{k: "utm_campaign", v: v}
}

func NewLighthouseURMSource(v string) LighthouseParam {
	return LighthouseParam{k: "utm_source", v: v}
}

func NewLighthouseCaptchaToken(v string) LighthouseParam {
	return LighthouseParam{k: "captchaToken", v: v}
}

func (p LighthouseParam) Key() string {
	return p.k
}

func (p LighthouseParam) Value() string {
	return p.v
}

type Environment struct {
	NetworkUserAgent string  `json:"networkUserAgent,omitempty"`
	HostUserAgent    string  `json:"hostUserAgent,omitempty"`
	BenchmarkIndex   float32 `json:"benchmarkIndex,omitempty"`
}

type RunWarning string

func (w RunWarning) Error() string {
	return string(w)
}

type ConfigSettings struct {
	EmulatedFormFactor string   `json:"emulatedFormFactor,omitempty"`
	FormFactor         string   `json:"formFactor,omitempty"`
	Locale             string   `json:"locale,omitempty"`
	FinalUrl           string   `json:"finalUrl,omitempty"`
	OnlyCategories     []string `json:"onlyCategories,omitempty"`
	Channel            string   `json:"channel,omitempty"`
}

type Audit struct {
	ID               string  `json:"id,omitempty"`
	Title            string  `json:"title,omitempty"`
	Description      string  `json:"description,omitempty"`
	Score            float32 `json:"score"`
	ScoreDisplayMode string  `json:"scoreDisplayMode,omitempty"`
}

type AuditRef struct {
	ID     string  `json:"id,omitempty"`
	Weight float32 `json:"weight,omitempty"`
	Group  string  `json:"group,omitempty"`
}

type Category struct {
	ID                string     `json:"id,omitempty"`
	Title             string     `json:"title,omitempty"`
	Description       string     `json:"description,omitempty"`
	Score             float32    `json:"score"`
	ManualDescription string     `json:"manualDescription,omitempty"`
	AuditRefs         []AuditRef `json:"auditRefs,omitempty"`
}

type CategoryGroup struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type Timing struct {
	Total float32 `json:"total,omitempty"`
}

type Screenshot struct {
	Data string `json:"data,omitempty"`
}

type ScreenshotNode struct {
	Height int `json:"height"`
	Left   int `json:"left"`
	Right  int `json:"right"`
	Width  int `json:"width"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
}

type FullPageScreenshot struct {
	Screenshot *Screenshot               `json:"screenshot,omitempty"`
	Nodes      map[string]ScreenshotNode `json:"nodes,omitempty"`
}

type Entity struct {
	Name           string   `json:"name,omitempty"`
	IsFirstParty   bool     `json:"isFirstParty,omitempty"`
	Category       string   `json:"category,omitempty"`
	IsUnrecognized bool     `json:"isUnrecognized,omitempty"`
	Origins        []string `json:"origins,omitempty"`
}

type LighthouseResult struct {
	RequestedUrl       string                   `json:"requestedUrl,omitempty"`
	FinalUrl           string                   `json:"finalUrl,omitempty"`
	MainDocumentUrl    string                   `json:"mainDocumentUrl,omitempty"`
	FinalDisplayedUrl  string                   `json:"finalDisplayedUrl,omitempty"`
	LighthouseVersion  string                   `json:"lighthouseVersion,omitempty"`
	UserAgent          string                   `json:"userAgent,omitempty"`
	FetchTime          *time.Time               `json:"fetchTime,omitempty"`
	Environment        *Environment             `json:"environment,omitempty"`
	RunWarnings        []RunWarning             `json:"runWarnings,omitempty"`
	ConfigSettings     *ConfigSettings          `json:"configSettings,omitempty"`
	Audits             map[string]Audit         `json:"audits,omitempty"`
	Categories         map[string]Category      `json:"categories,omitempty"`
	CategoryGroups     map[string]CategoryGroup `json:"categoryGroups,omitempty"`
	RuntimeError       *RuntimeError            `json:"runtimeError,omitempty"`
	Timing             *Timing                  `json:"timing,omitempty"`
	Entities           []Entity                 `json:"entities,omitempty"`
	FullPageScreenshot *FullPageScreenshot      `json:"fullPageScreenshot,omitempty"`
	internalError      error                    `json:"-"`
	m                  *sync.RWMutex            `json:"-"`
}

// NewLighthouseResult returns an initialized *LighthouseResult.
func NewLighthouseResult(u string) *LighthouseResult {

	r := new(LighthouseResult)

	r.RequestedUrl = u
	r.m = new(sync.RWMutex)

	return r
}

// RunLighthouse runs PageSpeed analysis on the page at the specified URL, and returns Lighthouse scores, a list of suggestions to make that page faster, and other information.
//
// The parameters must be specified in params. The url parameter is required!
// The Credential cred must be either *ApiKey or nil.
//
// The in the returned *LighthouseResult, the RequestedUrl field is always set to u.
// (Can be used as a key)
//
// If any error occurs, the returned error(s) are stored internally. Use the Errs() method to get the errors.
// The errors can comes from [http.Get], [json.Unmarshal] or can be type RuntimeError or RunWarning.
//
// API Reference: https://developers.google.com/speed/docs/insights/rest/v5/pagespeedapi/runpagespeed
func RunLighthouse(u string, cred Credential, params ...LighthouseParam) *LighthouseResult {

	r := NewLighthouseResult(u)

	r.Run(cred, params...)

	return r
}

// Run is similar to RunLighthouse. The URL is got from r.RequestURL.
func (r *LighthouseResult) Run(cred Credential, params ...LighthouseParam) {

	r.m.Lock()
	defer r.m.Unlock()

	uq := make(url.Values)

	uq.Add("url", r.RequestedUrl)

	if cred != nil {
		key, _ := cred.Token()
		uq.Add("key", key)
	}

	for i := range params {
		uq.Add(params[i].Key(), params[i].Value())
	}

	resp, err := http.Get("https://www.googleapis.com/pagespeedonline/v5/runPagespeed?" + uq.Encode())
	if err != nil {
		r.internalError = fmt.Errorf("failed HTTP Get: %w", err)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		r.internalError = NewRuntimeError(resp.StatusCode, fmt.Errorf("failed reading HTTP response: %w", err))
		return
	}

	v := struct {
		LighthouseResult *LighthouseResult `json:"lighthouseResult"`
		Err              *RuntimeError     `json:"error"`
	}{}

	err = json.Unmarshal(data, &v)
	if err != nil {
		r.internalError = NewRuntimeError(resp.StatusCode, fmt.Errorf("failed unmarshal response: %w", err))
		return
	}

	if v.LighthouseResult != nil {

		if v.LighthouseResult.m == nil {
			v.LighthouseResult.m = new(sync.RWMutex)
		}

		*r = *(*LighthouseResult)(v.LighthouseResult)
	}

	if v.Err != nil {
		r.internalError = *v.Err
	}
}

func (r *LighthouseResult) Errs() []error {

	r.m.RLock()
	defer r.m.RUnlock()

	var v []error = nil

	if r.internalError != nil {
		v = append(v, r.internalError)
	}

	if r.RuntimeError != nil {
		v = append(v, *r.RuntimeError)
	}

	for i := range r.RunWarnings {
		v = append(v, r.RunWarnings[i])

	}

	return v
}

func (r *LighthouseResult) SetErr(err error) {

	if r == nil {
		r = NewLighthouseResult("")
	}

	r.m.Lock()
	defer r.m.Unlock()

	r.internalError = err
}

func (r *LighthouseResult) Time() float32 {

	if r.Timing == nil {
		return 0
	}

	return r.Timing.Total
}

// RunConcurrentLighthouse use n number of workers to run RunLighthouse() concurrently.
//
// The Credential and the params are common across RunLighthouse functions.
func RunConcurrentLighthouse(ctx context.Context, n int, u []string, cred Credential, params ...LighthouseParam) []*LighthouseResult {

	s := semaphore.NewWeighted(int64(n))

	results := make([]*LighthouseResult, len(u))

	for i := range results {

		results[i] = NewLighthouseResult(u[i])

		if err := s.Acquire(ctx, 1); err != nil {
			results[i].SetErr(fmt.Errorf("failed to acquire semaphore: %w", err))
			continue
		}

		go func() {
			defer s.Release(1)
			results[i].Run(cred, params...)
			fmt.Printf("%s -> %.2fms\n", results[i].RequestedUrl, results[i].Time())
		}()
	}

	if err := s.Acquire(ctx, int64(n)); err != nil {
		return results
	}

	return results

}
