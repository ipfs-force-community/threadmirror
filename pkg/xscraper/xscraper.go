package xscraper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/xrate"
	"github.com/michimani/gotwi"
	"github.com/samber/lo"
)

//go:generate go tool oapi-codegen --config=types.cfg.yaml ./twitter-openapi/dist/docs/openapi-3.0.yaml

// default http client timeout
const DefaultClientTimeout = 10 * time.Second

var (
	// Twitter Web App (GraphQL API)
	TWITTER_WEB_APP_BEARER_TOKEN = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
	BASE_URL                     = lo.Must(url.Parse("https://x.com"))
)

type XScraper struct {
	realHeader        RealHeader
	xPrivateApiClient *http.Client
	rateLimiter       *xrate.Limiter
	logger            *slog.Logger

	// Gotwi client for media upload functionality
	gotwiClient *gotwi.Client

	LoginOpts     LoginOptions
	isLoggedIn    bool
	csrfToken     string
	loginMu       sync.Mutex
	initLoginOnce sync.Once
}

func New(loginOpts LoginOptions, logger *slog.Logger) (*XScraper, error) {
	var gotwiClient *gotwi.Client
	var err error
	if loginOpts.APIKey != "" && loginOpts.APIKeySecret != "" {
		gotwiClient, err = gotwi.NewClient(&gotwi.NewClientInput{
			AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
			APIKey:               loginOpts.APIKey,
			APIKeySecret:         loginOpts.APIKeySecret,
			OAuthToken:           loginOpts.AccessToken,
			OAuthTokenSecret:     loginOpts.AccessTokenSecret,
		})
		if err != nil {
			return nil, fmt.Errorf("create gotwi client: %w", err)
		}
	}

	return &XScraper{
		realHeader: RandRealHeader(),
		xPrivateApiClient: &http.Client{
			Jar:     lo.Must(cookiejar.New(nil)),
			Timeout: DefaultClientTimeout,
		},
		rateLimiter: xrate.NewLimiter(xrate.Every(1500*time.Millisecond), 1),

		logger:    logger,
		LoginOpts: loginOpts,

		gotwiClient: gotwiClient,
	}, nil
}

func (x *XScraper) Ready() bool {
	return x.rateLimiter.Tokens() >= 1
}

func (x *XScraper) WaitForReady(ctx context.Context) error {
	return x.rateLimiter.Wait(ctx)
}

func (x *XScraper) prepareRequest(req *http.Request) {
	req.Header.Set("X-Client-Transaction-Id", xClientTransactionID(req.Method, req.URL.Path))
	x.applyRealHeader(req)
	// x.setCSRFToken(req)
}

func (x *XScraper) applyRealHeader(req *http.Request) {
	x.realHeader.Apply(req)
	if req.Header.Get("Referer") == "" {
		req.Header.Set("Referer", "https://x.com/home")
	}
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", TWITTER_WEB_APP_BEARER_TOKEN)
	}
	if req.Header.Get("X-Csrf-Token") == "" && x.csrfToken != "" {
		req.Header.Set("X-Csrf-Token", x.csrfToken)
	}
}

func (x *XScraper) GetCookie(name string) (string, bool) {
	for _, cookie := range x.xPrivateApiClient.Jar.Cookies(BASE_URL) {
		if cookie.Name == name {
			return cookie.Value, true
		}
	}
	return "", false
}

func (x *XScraper) SetCookies(cs []*http.Cookie) {
	cookies := x.xPrivateApiClient.Jar.Cookies(BASE_URL)
	for _, c := range cs {
		if c.Domain == "" {
			c.Domain = "." + BASE_URL.Host
		}
		cookies = append(cookies, c)
	}
	x.xPrivateApiClient.Jar.SetCookies(BASE_URL, cookies)
}

func (x *XScraper) do(req *http.Request) (resp *http.Response, err error) {
	if err := x.WaitForReady(req.Context()); err != nil {
		return nil, fmt.Errorf("wait for ready: %w", err)
	}
	resp, err = x.xPrivateApiClient.Do(req)
	if err != nil {
		return
	}
	limit, err := strconv.Atoi(resp.Header.Get("x-rate-limit-limit"))
	if err == nil {
		limit = 0
	}
	reset, err := strconv.Atoi(resp.Header.Get("x-rate-limit-reset"))
	if err == nil {
		x.rateLimiter.ResetAt(time.Unix(int64(reset), 0), limit)
	}
	return resp, nil
}

type BadRequestError struct {
	StatusCode int
	Body       string
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("bad request: %d %s", e.StatusCode, e.Body)
}

func (x *XScraper) doJson(req *http.Request, target any) error {
	req.Header.Set("Content-Type", "application/json")
	resp, err := x.do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	fmt.Println(string(respBody))

	if http.StatusOK != resp.StatusCode {
		return &BadRequestError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}
	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("unmarshal response body %s: %w", respBody, err)
	}
	return nil
}

func (x *XScraper) GetGraphQL(ctx context.Context, endpoint string, params interface{ Query() url.Values }, target any) error {
	return x.DoGraphQL(ctx, http.MethodGet, endpoint, params, nil, target)
}

func (x *XScraper) PostGraphQL(ctx context.Context, endpoint string, req any, target any) error {
	return x.DoGraphQL(ctx, http.MethodPost, endpoint, nil, req, target)
}

var errNotLoggedIn = errors.New("not logged in")

func (x *XScraper) DoGraphQL(ctx context.Context, method, endpoint string, params interface{ Query() url.Values }, reqBody any, target any) error {
	// 初始化登录（只执行一次）
	x.initLoginOnce.Do(func() {
		ok, err := x.loadCookies(ctx)
		if err != nil {
			if x.logger != nil {
				x.logger.Error("failed to initialize login", "error", err)
			}
		}
		x.isLoggedIn = ok
	})

	for {
		// 检查登录状态并尝试登录
		if err := x.ensureLoggedIn(ctx); err != nil {
			return fmt.Errorf("ensure login: %w", err)
		}

		// 执行实际请求
		err := x.doGraphQL(ctx, method, endpoint, params, reqBody, target)
		if errors.Is(err, errNotLoggedIn) {
			// 如果收到未登录错误，标记为未登录状态，准备重试
			x.markNotLoggedIn()
			continue
		}

		return err
	}
}

func (x *XScraper) doGraphQL(ctx context.Context, method, endpoint string, params interface{ Query() url.Values }, reqBody any, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, BASE_URL.JoinPath(endpoint).String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if params != nil {
		query := params.Query()
		req.URL.RawQuery = query.Encode()
	}

	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(b))
	}

	x.prepareRequest(req)
	err = x.doJson(req, target)

	var berr *BadRequestError
	if err != nil {
		if errors.As(err, &berr) && berr.StatusCode == http.StatusUnauthorized || berr.StatusCode == http.StatusForbidden {
			return errNotLoggedIn
		}
		return err
	}
	return nil
}
