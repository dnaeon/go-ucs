package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dnaeon/go-ucs/mo"
	"golang.org/x/time/rate"
)

// RateLimit limits the number of requests per second that a Client
// can send to a remote Cisco UCS API endpoint using a rate.Limiter
// token bucket configured with the provided requests per seconds and
// burst. A request will wait for up to the given wait time.
type RateLimit struct {
	// RequestsPerSecond defines the allowed number of requests per second.
	RequestsPerSecond float64

	// Burst is the maximum burst size.
	Burst int

	// Wait defines the maximum time a request will wait for a token to be consumed.
	Wait time.Duration
}

// Config type contains the setting used by the Client.
type Config struct {
	// HttpClient is the HTTP client to use for sending requests.
	// If nil then we use http.DefaultClient for all requests.
	HttpClient *http.Client

	// Endpoint is the base URL to the remote Cisco UCS Manager endpoint.
	Endpoint string

	// Username to use when authenticating to the remote endpoint.
	Username string

	// Password to use when authenticating to the remote endpoint.
	Password string

	// RateLimit is used for limiting the number of requests per second
	// against the remote Cisco UCS API endpoint using a token bucket.
	RateLimit *RateLimit
}

// Client is used for interfacing with the remote Cisco UCS API endpoint.
type Client struct {
	config  *Config
	apiUrl  *url.URL
	limiter *rate.Limiter

	// Cookie is the authentication cookie currently in use.
	// It's value is set by the AaaLogin and AaaRefresh methods.
	Cookie string
}

// NewClient creates a new API client from the given config.
func NewClient(config Config) (*Client, error) {
	if config.HttpClient == nil {
		config.HttpClient = http.DefaultClient
	}

	baseUrl, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}

	apiUrl, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}

	var limiter *rate.Limiter
	if config.RateLimit != nil {
		rps := rate.Limit(config.RateLimit.RequestsPerSecond)
		limiter = rate.NewLimiter(rps, config.RateLimit.Burst)
	}

	client := &Client{
		config:  &config,
		apiUrl:  baseUrl.ResolveReference(apiUrl),
		limiter: limiter,
	}

	return client, nil
}

// Hostname returns the host portion of the remote UCS API endpoint without any port number.
func (c *Client) Hostname() string {
	return c.apiUrl.Host
}

// AaaLogin performs the initial authentication to the remote Cisco UCS API endpoint.
func (c *Client) AaaLogin(ctx context.Context) (*AaaLoginResponse, error) {
	req := AaaLoginRequest{
		InName:     c.config.Username,
		InPassword: c.config.Password,
	}

	var resp AaaLoginResponse
	if err := c.Request(ctx, req, &resp); err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.ToError()
	}

	// Set authentication cookie for future re-use when needed.
	c.Cookie = resp.OutCookie

	return &resp, nil
}

// AaaRefresh refreshes the current session by requesting a new authentication cookie.
func (c *Client) AaaRefresh(ctx context.Context) (*AaaRefreshResponse, error) {
	req := AaaRefreshRequest{
		InName:     c.config.Username,
		InPassword: c.config.Password,
		InCookie:   c.Cookie,
	}

	var resp AaaRefreshResponse
	if err := c.Request(ctx, req, &resp); err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.ToError()
	}

	// Set new authentication cookie
	c.Cookie = resp.OutCookie

	return &resp, nil
}

// AaaKeepAlive sends a request to keep the current session active using the same cookie.
func (c *Client) AaaKeepAlive(ctx context.Context) (*AaaKeepAliveResponse, error) {
	req := AaaKeepAliveRequest{
		Cookie: c.Cookie,
	}

	var resp AaaKeepAliveResponse
	if err := c.Request(ctx, req, &resp); err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.ToError()
	}

	return &resp, nil
}

// AaaLogout invalidates the current client session.
func (c *Client) AaaLogout(ctx context.Context) (*AaaLogoutResponse, error) {
	req := AaaLogoutRequest{
		InCookie: c.Cookie,
	}

	var resp AaaLogoutResponse
	if err := c.Request(ctx, req, &resp); err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.ToError()
	}

	c.Cookie = ""

	return &resp, nil
}

// doRequest sends a request to the remote Cisco UCS API endpoint.
func (c *Client) doRequest(ctx context.Context, in, out interface{}) error {
	data, err := xmlMarshalWithSelfClosingTags(in)
	if err != nil {
		return err
	}

	r, err := http.NewRequest("POST", c.apiUrl.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req := r.WithContext(ctx)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.config.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return xml.Unmarshal(body, &out)
}

// Request sends a POST request to the remote Cisco UCS API endpoint.
func (c *Client) Request(ctx context.Context, in, out interface{}) error {
	// Rate limit requests if we are using a limiter
	if c.limiter != nil {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.RateLimit.Wait)
		defer cancel()
		if err := c.limiter.Wait(ctxWithTimeout); err != nil {
			return err
		}
	}

	return c.doRequest(ctx, in, out)
}

// RequestNow sends a POST request to the remote Cisco UCS API endpoint immediately.
// This bypasses any rate limiter configuration that may be used and is
// meant to be used for priority requests, e.g. refreshing a token, logging out, etc.
func (c *Client) RequestNow(ctx context.Context, in, out interface{}) error {
	return c.doRequest(ctx, in, out)
}

// ConfigResolveDn retrieves a single managed object for a specified DN.
func (c *Client) ConfigResolveDn(ctx context.Context, in ConfigResolveDnRequest, out mo.Any) error {
	var resp ConfigResolveDnResponse
	if err := c.Request(ctx, in, &resp); err != nil {
		return err
	}

	if resp.IsError() {
		return resp.ToError()
	}

	// The requested managed object is contained within the inner XML document,
	// which we need to unmarshal first into the given concrete type.
	return xml.Unmarshal(resp.OutConfig.Inner, &out)

}

// ConfigResolveDns retrieves managed objects for a specified list of DNs.
func (c *Client) ConfigResolveDns(ctx context.Context, in ConfigResolveDnsRequest, out mo.Any) (*ConfigResolveDnsResponse, error) {
	var resp ConfigResolveDnsResponse
	if err := c.Request(ctx, in, &resp); err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, resp.ToError()
	}

	inner, err := xml.Marshal(resp.OutConfigs)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(inner, &out); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ConfigResolveClass retrieves managed objects of the specified class.
func (c *Client) ConfigResolveClass(ctx context.Context, in ConfigResolveClassRequest, out mo.Any) error {
	var resp ConfigResolveClassResponse
	if err := c.Request(ctx, in, &resp); err != nil {
		return err
	}

	if resp.IsError() {
		return resp.ToError()
	}

	inner, err := xml.Marshal(resp.OutConfigs)
	if err != nil {
		return err
	}

	// The requested managed objects are contained within the inner XML document,
	// which we need to unmarshal first into the given concrete type.
	return xml.Unmarshal(inner, &out)
}

// ConfigResolveClasses retrieves managed objects from the specified list of classes.
func (c *Client) ConfigResolveClasses(ctx context.Context, in ConfigResolveClassesRequest, out mo.Any) error {
	var resp ConfigResolveClassesResponse
	if err := c.Request(ctx, in, &resp); err != nil {
		return err
	}

	if resp.IsError() {
		return resp.ToError()
	}

	inner, err := xml.Marshal(resp.OutConfigs)
	if err != nil {
		return err
	}

	// The requested managed objects are contained within the inner XML document,
	// which we need to unmarshal first into the given concrete type.
	return xml.Unmarshal(inner, &out)
}

// ConfigResolveChildren retrieves children of managed objects under a specified DN.
func (c *Client) ConfigResolveChildren(ctx context.Context, in ConfigResolveChildrenRequest, out mo.Any) error {
	var resp ConfigResolveChildrenResponse

	if err := c.Request(ctx, in, &resp); err != nil {
		return err
	}

	if resp.IsError() {
		return resp.ToError()
	}

	inner, err := xml.Marshal(resp.OutConfigs)
	if err != nil {
		return err
	}

	// The requested managed objects are contained within the inner XML document,
	// which we need to unmarshal first into the given concrete type.
	return xml.Unmarshal(inner, &out)
}
