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
	LighthouseCategoryAccessibility = LighthouseCategory("ACCESSIBILITY")
	LighthouseCategoryBestPractices = LighthouseCategory("BEST_PRACTICES")
	LighthouseCategoryPerformance   = LighthouseCategory("PERFORMANCE")
	LighthouseCategorySEO           = LighthouseCategory("SEO")

	// Shorthand to all PagespeedCategory
	LighthouseCategoryAll = []LighthouseParam{LighthouseCategoryAccessibility, LighthouseCategoryBestPractices, LighthouseCategoryPerformance, LighthouseCategorySEO}
)

// Possible values for strategy paramater
var (
	LighthouseStrategyDesktop = LighthouseStrategy("dektop")
	LighthouseStrategyMobile  = LighthouseStrategy("mobile")
)

var (

	// Invalid domain
	ErrLighthouseFailedDocumentRequest = &GoogleError{
		Message: "Lighthouse returned error: FAILED_DOCUMENT_REQUEST. Lighthouse was unable to reliably load the page you requested. Make sure you are testing the correct URL and that the server is properly responding to all requests. (Details: net::ERR_CONNECTION_FAILED)",
		Domain:  "lighthouse",
		Reason:  "lighthouseUserError",
	}

	// Too much request
	ErrLighthouseUnprocessable = &GoogleError{
		Message: "Unable to process request. Please wait a while and try again.",
		Domain:  "global",
		Reason:  "internalError",
	}

	ErrLighthouseInvalidKey = &GoogleError{
		Message: "API key not valid. Please pass a valid API key.",
		Domain:  "global",
		Reason:  "badRequest",
	}

	ErrLighthouseInvalidCategory = &GoogleError{
		Message: `^Invalid value at 'category' \(type\.googleapis\.com/google\.chrome\.pagespeedonline\.v5\.PagespeedonlinePagespeedapiRunpagespeedRequest\.Category\), .*$`,
		Reason:  "invalid",
	}

	ErrLighthouseInvalidStrategy = &GoogleError{
		Message: `^Invalid value at 'strategy' \(type\.googleapis\.com/google\.chrome\.pagespeedonline\.v5\.PagespeedonlinePagespeedapiRunpagespeedRequest\.Strategy\), .*$`,
		Reason:  "invalid",
	}

	ErrLighthouseInvalidUrl = &GoogleError{
		Message:      `^Invalid value '.*'\. Values must match the following regular expression: '\(\?i\)\(url:\|origin:\)\?http\(s\)\?://\.\*'$`,
		Domain:       "gdata.CoreErrorDomain",
		Reason:       "INVALID_PARAMETER",
		Location:     "url",
		LocationType: "other",
	}

	// Rate limit
	ErrLighthouseRateLimitExceeded = &GoogleError{
		Message: `^Quota exceeded for quota metric 'Queries' and limit 'Queries per minute' of service 'pagespeedonline\.googleapis\.com' for consumer '.*'\.$`,
		Domain:  "global",
		Reason:  "rateLimitExceeded",
	}
)

type LighthouseError struct {
	URL string
	Err error
}

func NewLighthouseError(url string, err error) *LighthouseError {

	return &LighthouseError{URL: url, Err: err}
}

func (e *LighthouseError) Error() string {
	return fmt.Sprintf("\"%s\": \"%s\"", e.URL, e.Err)
}

func (e *LighthouseError) Unwrap() error {
	return e.Err
}

// LighthouseParam stores a single request param as a key/value pair.
type LighthouseParam struct {
	k, v string
}

func LighthouseCategory(v string) LighthouseParam {
	return LighthouseParam{k: "category", v: v}
}

func LighthouseLocale(v string) LighthouseParam {
	return LighthouseParam{k: "locale", v: v}
}

func LighthouseStrategy(v string) LighthouseParam {
	return LighthouseParam{k: "strategy", v: v}
}

func LighthouseUTMCampaign(v string) LighthouseParam {
	return LighthouseParam{k: "utm_campaign", v: v}
}

func LighthouseURMSource(v string) LighthouseParam {
	return LighthouseParam{k: "utm_source", v: v}
}

func LighthouseCaptchaToken(v string) LighthouseParam {
	return LighthouseParam{k: "captchaToken", v: v}
}

func (p LighthouseParam) Key() string {
	return p.k
}

func (p LighthouseParam) Value() string {
	return p.v
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

type LighthouseResult struct {
	requestedUrl   *url.URL                  // The original requested url.
	finalUrl       *url.URL                  // The final resolved url that was audited.
	fetchTime      time.Time                 // The time that this run was fetched. (eg.: "2024-07-29T16:25:29.029Z")
	runWarnings    []error                   //
	audits         map[string]*Audit         // An object containing the results of the audits.
	categories     map[string]*Category      // Map of categories in the LHR.
	categoryGroups map[string]*CategoryGroup //

	timing time.Duration // The total duration of Lighthouse's run.
}

// createLighthouseURL returns the complete URL that can be passed to http.Get().
//
// Appends the request parameters to the API endpoint.
//
// If any error returned, that comes from Credential cred.
func CreateLighthouseURL(u string, cred Credential, params ...LighthouseParam) (string, error) {

	// Create the query string
	uq := make(url.Values)

	uq.Add("url", u)

	// Get token string
	if cred != nil {
		key, err := cred.Token()
		if err != nil {
			return "", fmt.Errorf("token error: %w", err)
		}

		uq.Add("key", key)

	}

	for i := range params {
		uq.Add(params[i].Key(), params[i].Value())
	}

	v := "https://www.googleapis.com/pagespeedonline/v5/runPagespeed?" + uq.Encode()

	return v, nil

}

// LighthouseResultFromResponse reads the LighthouseResult from an *http.Response (eg.: http.Response) and unmarshals it.
func LighthouseResultFromResponse(r *http.Response) (*LighthouseResult, error) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	v := struct {
		LighthouseResult *LighthouseResult `json:"lighthouseResult"`
	}{}

	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	return v.LighthouseResult, nil
}

// RunLighthouse runs PageSpeed analysis on the page at the specified URL, and returns Lighthouse scores, a list of suggestions to make that page faster, and other information.
//
// The url parameter is required!
// The parameters must be specified in params.
// The Credential cred must be either *ApiKey or nil.
//
// If any error occurs, the returned error is always *LighthouseError.
// Errors comes from other packages are wrapped in the *LighthouseError (eg.: [http.Get], [json.Unmarshal]).
//
// API Reference: https://developers.google.com/speed/docs/insights/rest/v5/pagespeedapi/runpagespeed
func RunLighthouse(u string, cred Credential, params ...LighthouseParam) (*LighthouseResult, error) {

	getUrl, err := CreateLighthouseURL(u, cred, params...)
	if err != nil {
		return nil, NewLighthouseError(u, err)
	}

	resp, err := http.Get(getUrl)
	if err != nil {
		return nil, NewLighthouseError(u, err)
	}
	defer resp.Body.Close()

	// Error
	if resp.StatusCode != 200 {

		gerr, err := ErrorFromResponse(resp)
		if err != nil {
			return nil, NewLighthouseError(u, err)
		}

		return nil, NewLighthouseError(u, gerr)
	}

	r, err := LighthouseResultFromResponse(resp)
	if err != nil {
		return nil, NewLighthouseError(u, err)
	}

	return r, nil
}

func (r *LighthouseResult) UnmarshalJSON(data []byte) error {

	v := struct {
		RequestedUrl   string                    `json:"requestedUrl"`
		FinalUrl       string                    `json:"finalUrl"`
		FetchTime      string                    `json:"fetchTime"`
		RunWarnings    []string                  `json:"runWarnings"`
		Audits         map[string]*Audit         `json:"audits"`
		Categories     map[string]*Category      `json:"categories"`
		CategoryGroups map[string]*CategoryGroup `json:"categoryGroups"`
		RuntimeError   *Error                    `json:"runtimeError"`
		Timing         struct {
			Total float64 `json:"total"`
		} `json:"timing"`
	}{}

	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	for i := range v.RunWarnings {
		r.runWarnings = append(r.runWarnings, fmt.Errorf(v.RunWarnings[i]))
	}

	if v.RuntimeError != nil {
		v.RuntimeError.Errors = append(v.RuntimeError.Errors, r.runWarnings...)

		return v.RuntimeError
	}

	r.requestedUrl, err = url.Parse(v.RequestedUrl)
	if err != nil {
		return fmt.Errorf("invalid requestedUrl: %w", err)
	}

	r.finalUrl, err = url.Parse(v.FinalUrl)
	if err != nil {
		return fmt.Errorf("invalid finalUrl: %w", err)
	}

	r.fetchTime, err = time.Parse(time.RFC3339, v.FetchTime)
	if err != nil {
		return fmt.Errorf("invalid fetchTime: %w", err)
	}

	r.audits = v.Audits

	r.categories = v.Categories

	r.categoryGroups = v.CategoryGroups

	r.timing = time.Duration(v.Timing.Total * 1_000_000)

	return nil
}

// RequestedURL returns the original requested url.
func (r *LighthouseResult) RequestedURL() *url.URL {
	return r.requestedUrl
}

// FinalURL returns the final resolved url that was audited.
func (r *LighthouseResult) FinalURL() *url.URL {
	return r.finalUrl
}

// FetchTime returns the time that this run was fetched.
func (r *LighthouseResult) FetchTime() time.Time {
	return r.fetchTime
}

// RunWarnings returns warnings (non-fatal errors) coming from the PageSpeed API.
func (r *LighthouseResult) RunWarnings() []error {
	return r.runWarnings
}

// Audit returns the audit with the given name.
//
// If audit not found, returns nil.
func (r *LighthouseResult) Audit(name string) *Audit {
	return r.audits[name]
}

// Audits returns the names of the available audits.
func (r *LighthouseResult) Audits() []string {

	v := make([]string, 0, len(r.audits))

	for k := range r.audits {
		v = append(v, k)
	}

	return v
}

// Category returns the category with the given name.
//
// Possible categories:
//
//	"performance"
//	"accessibility"
//
// If category not found, returns nil.
func (r *LighthouseResult) Category(name string) *Category {
	return r.categories[name]
}

// Categories returns the names of the available categories.
func (r *LighthouseResult) Categories() []string {

	v := make([]string, 0, len(r.categories))

	for k := range r.categories {
		v = append(v, k)
	}

	return v
}

// CategoryGroup returns the category group with the given name.
//
// If category group not found, returns nil.
func (r *LighthouseResult) CategoryGroup(name string) *CategoryGroup {
	return r.categoryGroups[name]
}

// CategoryGroups returns the names of the available category groups.
func (r *LighthouseResult) CategoryGroups() []string {

	v := make([]string, 0, len(r.categoryGroups))

	for k := range r.categoryGroups {
		v = append(v, k)
	}

	return v
}

func (r *LighthouseResult) Timing() time.Duration {
	return r.timing
}

// Score returns the score of the category.
//
// If "average" is used as category, returns the average score of the available categories.
// If "total" is used as category, returns the total score (adds the scores of the available categories).
//
// If category not found returns -1.
func (r *LighthouseResult) Score(category string) int {

	if len(r.categories) == 0 {
		return 0
	}

	switch category {
	case "performance", "accessibility", "best-practices", "seo":
		c := r.Category(category)
		if c == nil {
			return -1
		}
		return int(c.Score * 100)

	case "average":
		var s float32
		for k := range r.categories {
			s += r.categories[k].Score
		}
		return int(s*100) / len(r.categories)

	case "total":
		var s float32
		for k := range r.categories {
			s += r.categories[k].Score
		}
		return int(s * 100)

	default:
		return -1
	}
}

// RunConcurrentLighthouse use n number of workers to run RunLighthouse() concurrently.
//
// The Credential and the params are common across RunLighthouse functions.
func RunConcurrentLighthouse(ctx context.Context, n int, u []string, cred Credential, params ...LighthouseParam) ([]*LighthouseResult, []error) {

	s := semaphore.NewWeighted(int64(n))
	m := new(sync.Mutex)

	r := make([]*LighthouseResult, 0, len(u))
	e := make([]error, 0, len(u))

	for i := range u {

		if err := s.Acquire(ctx, 1); err != nil {
			e = append(e, NewLighthouseError(u[i], err))
			continue
		}

		go func() {
			defer s.Release(1)

			if lr, err := RunLighthouse(u[i], cred, params...); err != nil {
				m.Lock()
				e = append(e, err)
				m.Unlock()
			} else {
				m.Lock()
				r = append(r, lr)
				m.Unlock()
			}
		}()
	}

	if err := s.Acquire(ctx, int64(n)); err != nil {
		return nil, []error{NewLighthouseError("", err)}
	}

	return r, e

}
