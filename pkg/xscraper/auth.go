package xscraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/samber/lo"
)

type privLoginState int8

const (
	privLoginStateInitPrivateApi privLoginState = iota
	privLoginStateLoginJsInstrumentationSubtask
	privLoginStateLoginEnterUserIdentifierSSO
	privLoginStateLoginEnterAlternateIdentifierSubtask
	privLoginStateLoginEnterPassword
	privLoginStateAccountDuplicationCheck
	privLoginStateLoginTwoFactorAuthChallenge
	privLoginStateLoginAcid
	privLoginStateLoginSuccessSubtask
	privLoginStateDenyLoginSubtask
	privLoginStateLoggedIn
	privLoginStateUnknown
)

type (
	flowTaskRequest struct {
		FlowToken       string           `json:"flow_token,omitempty"`
		SubtaskInputs   []map[string]any `json:"subtask_inputs"`
		InputFlowData   map[string]any   `json:"input_flow_data,omitempty"`
		SubtaskVersions map[string]any   `json:"subtask_versions,omitempty"`
	}

	flowResponse struct {
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors,omitempty"`
		FlowToken string `json:"flow_token"`
		Status    string `json:"status"`
		Subtasks  []struct {
			SubtaskID string `json:"subtask_id"`
		} `json:"subtasks"`
	}
)

type LoginOptions struct {
	LoadCookies       func(ctx context.Context) ([]*http.Cookie, error)
	SaveCookies       func(ctx context.Context, cookies []*http.Cookie) error
	Username          string
	Password          string
	Email             string
	APIKey            string
	APIKeySecret      string
	AccessToken       string
	AccessTokenSecret string
}

func (x *XScraper) loadCookies(ctx context.Context) (bool, error) {
	cookies, err := x.loginOpts.LoadCookies(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to load cookies: %w", err)
	}
	if len(cookies) != 0 {
		x.loginMu.Lock()
		defer x.loginMu.Unlock()

		x.xPrivateApiClient.Jar.SetCookies(BASE_URL, cookies)
		if csrfToken, ok := x.GetCookie("ct0"); ok {
			x.csrfToken = csrfToken
		}
		return true, nil
	}
	return false, nil
}

func (x *XScraper) checkLoggedIn() bool {
	x.loginMu.Lock()
	defer x.loginMu.Unlock()

	if !x.isLoggedIn {
		return false
	}

	cookies := x.xPrivateApiClient.Jar.Cookies(BASE_URL)
	hasAuthToken := false
	hasCsrfToken := false
	for _, cookie := range cookies {
		switch cookie.Name {
		case "auth_token":
			hasAuthToken = true
		case "ct0":
			hasCsrfToken = true
		}
	}
	x.isLoggedIn = hasAuthToken && hasCsrfToken
	return x.isLoggedIn
}

// ensureLoggedIn 确保当前处于已登录状态
func (x *XScraper) ensureLoggedIn(ctx context.Context) error {
	if x.checkLoggedIn() {
		return nil
	}

	x.loginMu.Lock()
	defer x.loginMu.Unlock()

	authHandler := &authHandler{
		scraper: x,
	}
	if err := authHandler.tryLogin(ctx); err != nil {
		return err
	}
	x.isLoggedIn = true
	return nil
}

func (x *XScraper) markNotLoggedIn() {
	x.loginMu.Lock()
	defer x.loginMu.Unlock()
	x.isLoggedIn = false
	x.xPrivateApiClient.Jar = lo.Must(cookiejar.New(nil))
	x.csrfToken = ""
}

type authHandler struct {
	scraper    *XScraper
	csrfToken  string
	guestToken string
}

func (a *authHandler) prepareRequest(req *http.Request) {
	a.scraper.applyRealHeader(req)
	if a.csrfToken != "" {
		req.Header.Set("X-Csrf-Token", a.csrfToken)
	}
	if a.guestToken != "" {
		req.Header.Set("X-Guest-Token", a.guestToken)
	}
	req.Header.Set("X-Twitter-Active-User", "yes")
	req.Header.Set("2", "n")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("X-Client-Transaction-Id", xClientTransactionID(req.Method, req.URL.Path))
}

func (a *authHandler) tryLogin(ctx context.Context) error { //nolint:unparam
	var (
		nextState     = privLoginStateInitPrivateApi
		nextFlowToken string
		nextSubtaskID string
		err           error
	)

	for {
		switch nextState {
		case privLoginStateInitPrivateApi:
			nextFlowToken, nextSubtaskID, err = a.initPrivateApiLogin(ctx)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateLoginJsInstrumentationSubtask:
			nextFlowToken, nextSubtaskID, err = a.handleLoginJsInstrumentationSubtask(ctx, nextFlowToken)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateLoginEnterUserIdentifierSSO:
			nextFlowToken, nextSubtaskID, err = a.handleLoginEnterUserIdentifierSSO(ctx, nextFlowToken, a.scraper.loginOpts.Username)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateLoginEnterAlternateIdentifierSubtask:
			nextFlowToken, nextSubtaskID, err = a.handleLoginEnterAlternateIdentifierSubtask(ctx, nextFlowToken, a.scraper.loginOpts.Email)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateLoginEnterPassword:
			nextFlowToken, nextSubtaskID, err = a.handleLoginEnterPassword(ctx, nextFlowToken, a.scraper.loginOpts.Password)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateAccountDuplicationCheck:
			nextFlowToken, nextSubtaskID, err = a.handleAccountDuplicationCheck(ctx, nextFlowToken)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateLoginTwoFactorAuthChallenge:
			nextFlowToken, nextSubtaskID, err = a.handleLoginTwoFactorAuthChallenge(ctx, nextFlowToken)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateLoginAcid:
			nextFlowToken, nextSubtaskID, err = a.handleLoginAcid(ctx, nextFlowToken, &a.scraper.loginOpts.Email)
			if err != nil {
				return err
			}
			nextState = subtaskIDToPrivLoginState(nextSubtaskID)
		case privLoginStateLoginSuccessSubtask:
			nextFlowToken, _, err = a.handleLoginSuccessSubtask(ctx, nextFlowToken)
			if err != nil {
				return err
			}
			nextState = privLoginStateLoggedIn
		case privLoginStateDenyLoginSubtask:
			return fmt.Errorf("authentication error: DenyLoginSubtask")
		case privLoginStateLoggedIn:
			err = a.scraper.loginOpts.SaveCookies(ctx, a.scraper.xPrivateApiClient.Jar.Cookies(BASE_URL))
			return err

		case privLoginStateUnknown:
			return fmt.Errorf("unknown state")
		}
		// sleep 0.5s to 1.5s
		time.Sleep(time.Duration(500+rand.IntN(1000)) * time.Millisecond)
	}
}

func (a *authHandler) requestHashflags(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.x.com/1.1/hashflags.json", nil)
	if err != nil {
		return fmt.Errorf("create hashflags request: %w", err)
	}
	a.prepareRequest(req)
	resp, err := a.scraper.do(req)
	if err != nil {
		return fmt.Errorf("send hashflags request: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (a *authHandler) requestGuestToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.x.com/1.1/guest/activate.json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	a.prepareRequest(req)
	type guestTokenResponse struct {
		GuestToken string `json:"guest_token"`
	}
	var respData guestTokenResponse
	err = a.scraper.doJson(req, &respData)
	if err != nil {
		return "", fmt.Errorf("failed to send http request: %w", err)
	}

	return respData.GuestToken, nil
}

func (a *authHandler) simulateXWebRequest(ctx context.Context) (err error) {
	// access to https://x.com/i/flow/login to get cookies
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://x.com/i/flow/login", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	a.prepareRequest(req)
	resp, err := a.scraper.do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if _, ok := a.scraper.GetCookie("gt"); !ok {
		a.guestToken, err = a.requestGuestToken(ctx)
		if err != nil {
			return fmt.Errorf("request guest token: %w", err)
		}
		a.scraper.SetCookies([]*http.Cookie{
			{
				Name:   "gt",
				Value:  a.guestToken,
				Domain: "." + BASE_URL.Host,
			},
		})
	}

	if err := a.requestHashflags(ctx); err != nil {
		return fmt.Errorf("request hashflags: %w", err)
	}

	return
}

func (a *authHandler) initPrivateApiLogin(ctx context.Context) (flowToken, subtaskID string, err error) {
	err = a.simulateXWebRequest(ctx)
	if err != nil {
		return "", "", fmt.Errorf("simulate x web request: %w", err)
	}

	return a.executeFlowTask(ctx, flowTaskRequest{
		InputFlowData: map[string]any{
			"flow_context": map[string]any{
				"debug_overrides": map[string]any{},
				"start_location": map[string]any{
					"location": "splash_screen",
				},
			},
		},
		SubtaskVersions: map[string]any{
			"action_list":                          2,
			"alert_dialog":                         1,
			"app_download_cta":                     1,
			"check_logged_in_account":              1,
			"choice_selection":                     3,
			"contacts_live_sync_permission_prompt": 0,
			"cta":                                  7,
			"email_verification":                   2,
			"end_flow":                             1,
			"enter_date":                           1,
			"enter_email":                          2,
			"enter_password":                       5,
			"enter_phone":                          2,
			"enter_recaptcha":                      1,
			"enter_text":                           5,
			"enter_username":                       2,
			"generic_urt":                          3,
			"in_app_notification":                  1,
			"interest_picker":                      3,
			"js_instrumentation":                   1,
			"menu_dialog":                          1,
			"notifications_permission_prompt":      2,
			"open_account":                         2,
			"open_home_timeline":                   1,
			"open_link":                            1,
			"phone_verification":                   4,
			"privacy_options":                      1,
			"security_key":                         3,
			"select_avatar":                        4,
			"select_banner":                        2,
			"settings_list":                        7,
			"show_code":                            1,
			"sign_up":                              2,
			"sign_up_review":                       4,
			"tweet_selection_urt":                  1,
			"update_users":                         1,
			"upload_media":                         1,
			"user_recommendations_list":            4,
			"user_recommendations_urt":             1,
			"wait_spinner":                         3,
			"web_modal":                            1,
		},
	})
}

func (a *authHandler) handleLoginJsInstrumentationSubtask(ctx context.Context, flowToken string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken: flowToken,
		SubtaskInputs: []map[string]any{{
			"subtask_id": "LoginJsInstrumentationSubtask",
			"js_instrumentation": map[string]any{
				"link": "next_link",
			},
		}},
	})
}

func (a *authHandler) handleLoginEnterUserIdentifierSSO(ctx context.Context, flowToken, username string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken: flowToken,
		SubtaskInputs: []map[string]any{{
			"subtask_id": "LoginEnterUserIdentifierSSO",
			"settings_list": map[string]any{
				"setting_responses": []map[string]any{{
					"key": "user_identifier",
					"response_data": map[string]any{
						"text_data": map[string]any{"result": username},
					},
				}},
				"link": "next_link",
			},
		}},
	})
}

func (a *authHandler) handleLoginEnterAlternateIdentifierSubtask(ctx context.Context, flowToken, email string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken: flowToken,
		SubtaskInputs: []map[string]any{{
			"subtask_id": "LoginEnterAlternateIdentifierSubtask",
			"enter_text": map[string]any{
				"text": email,
				"link": "next_link",
			},
		}},
	})
}

func (a *authHandler) handleLoginEnterPassword(ctx context.Context, flowToken, password string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken: flowToken,
		SubtaskInputs: []map[string]any{{
			"subtask_id": "LoginEnterPassword",
			"enter_password": map[string]any{
				"password": password,
				"link":     "next_link",
			},
		}},
	})
}

func (a *authHandler) handleAccountDuplicationCheck(ctx context.Context, flowToken string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken: flowToken,
		SubtaskInputs: []map[string]any{{
			"subtask_id": "AccountDuplicationCheck",
			"check_logged_in_account": map[string]any{
				"link": "AccountDuplicationCheck_false",
			},
		}},
	})
}

func (a *authHandler) handleLoginTwoFactorAuthChallenge(ctx context.Context, flowToken string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken: flowToken,
		SubtaskInputs: []map[string]any{{
			"subtask_id": "LoginTwoFactorAuthChallenge",
		}},
	})
}

func (a *authHandler) handleLoginAcid(ctx context.Context, flowToken string, email *string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken: flowToken,
		SubtaskInputs: []map[string]any{{
			"subtask_id": "LoginAcid",
			"enter_text": map[string]any{
				"text": email,
				"link": "next_link",
			},
		}},
	})
}

func (a *authHandler) handleLoginSuccessSubtask(ctx context.Context, flowToken string) (nextFlowToken, subtaskID string, err error) {
	return a.executeFlowTask(ctx, flowTaskRequest{
		FlowToken:     flowToken,
		SubtaskInputs: []map[string]any{},
	})
}

func (a *authHandler) executeFlowTask(ctx context.Context, flowTaskRequest flowTaskRequest) (flowToken string, subtaskID string, err error) {
	var onboardingTaskUrl string
	if flowTaskRequest.FlowToken == "" {
		onboardingTaskUrl = "https://api.x.com/1.1/onboarding/task.json?flow_name=login"
	} else {
		onboardingTaskUrl = "https://api.x.com/1.1/onboarding/task.json"
	}

	body, err := json.Marshal(flowTaskRequest)
	if err != nil {
		return "", "", fmt.Errorf("marshal flow task request %#v: %w", flowTaskRequest, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, onboardingTaskUrl, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	a.prepareRequest(req)
	var flowResp flowResponse
	err = a.scraper.doJson(req, &flowResp)
	a.scraper.logger.Debug("execute flow task", "flowTaskRequest", flowTaskRequest, "flowResp", flowResp, "request header", req.Header, "err", err)
	if err != nil {
		return "", "", fmt.Errorf("execute flow task: %w", err)
	}

	if ct0, ok := a.scraper.GetCookie("ct0"); ok {
		a.csrfToken = ct0
	}

	if len(flowResp.Errors) > 0 {
		return "", "", fmt.Errorf("authentication error(%d): %s", flowResp.Errors[0].Code, flowResp.Errors[0].Message)
	}

	flowToken = flowResp.FlowToken
	if len(flowResp.Subtasks) > 0 {
		subtaskID = flowResp.Subtasks[0].SubtaskID
	}

	return
}

func subtaskIDToPrivLoginState(subtaskID string) privLoginState {
	switch subtaskID {
	case "LoginJsInstrumentationSubtask":
		return privLoginStateLoginJsInstrumentationSubtask
	case "LoginEnterUserIdentifierSSO":
		return privLoginStateLoginEnterUserIdentifierSSO
	case "LoginEnterAlternateIdentifierSubtask":
		return privLoginStateLoginEnterAlternateIdentifierSubtask
	case "LoginEnterPassword":
		return privLoginStateLoginEnterPassword
	case "AccountDuplicationCheck":
		return privLoginStateAccountDuplicationCheck
	case "LoginTwoFactorAuthChallenge":
		return privLoginStateLoginTwoFactorAuthChallenge
	case "LoginAcid":
		return privLoginStateLoginAcid
	case "LoginSuccessSubtask":
		return privLoginStateLoginSuccessSubtask
	case "DenyLoginSubtask":
		return privLoginStateDenyLoginSubtask
	default:
		return privLoginStateUnknown
	}
}
