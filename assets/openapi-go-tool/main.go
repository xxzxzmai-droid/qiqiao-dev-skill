package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	BaseURL            string `json:"baseUrl"`
	CorpID             string `json:"corpId"`
	Secret             string `json:"secret"`
	Account            string `json:"account"`
	ApplicationID      string `json:"applicationId"`
	FormModelID        string `json:"formModelId"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	UseProxy           bool   `json:"useProxy"`
	TimeoutSeconds     int    `json:"timeoutSeconds"`
}

type APIEnvelope struct {
	Code any             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type ErrorResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

type RecordRequest struct {
	ID          string         `json:"id"`
	LoginUserID string         `json:"loginUserId"`
	Version     int            `json:"version"`
	Variables   map[string]any `json:"variables"`
}

type FormField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
}

type TestStep struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Code    any    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Shape   any    `json:"shape,omitempty"`
}

type QiqiaoClient struct {
	cfg    Config
	client *http.Client

	mu        sync.Mutex
	token     string
	tokenTime time.Time
}

func main() {
	var (
		configPath       string
		qqkfPath         string
		mode             string
		listen           string
		variablesJSON    string
		pageSize         int
		deleteAfterSmoke bool
		uploadFilePath   string
	)

	cfg := Config{TimeoutSeconds: 20}

	flag.StringVar(&mode, "mode", "serve", "mode: serve, probe, schema, crud-smoke, full-smoke, design-probe")
	flag.StringVar(&listen, "listen", "127.0.0.1:8787", "listen address for serve mode")
	flag.StringVar(&configPath, "config", getenv("QIQIAO_CONFIG", ""), "JSON config path")
	flag.StringVar(&qqkfPath, "qqkf", getenv("QIQIAO_QQKF", ""), "qqkf.txt path")
	flag.StringVar(&cfg.BaseURL, "base-url", getenv("QIQIAO_BASE_URL", cfg.BaseURL), "Qiqiao OpenAPI base URL")
	flag.StringVar(&cfg.CorpID, "corp-id", getenv("QIQIAO_CORP_ID", ""), "corp id")
	flag.StringVar(&cfg.Secret, "secret", getenv("QIQIAO_SECRET", ""), "tenant secret")
	flag.StringVar(&cfg.Account, "account", getenv("QIQIAO_ACCOUNT", ""), "admin account")
	flag.StringVar(&cfg.ApplicationID, "app-id", getenv("QIQIAO_APPLICATION_ID", ""), "application id")
	flag.StringVar(&cfg.FormModelID, "form-id", getenv("QIQIAO_FORM_MODEL_ID", ""), "form model id")
	flag.BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-verify", getenvBool("QIQIAO_INSECURE_SKIP_VERIFY", false), "skip HTTPS certificate verification")
	flag.BoolVar(&cfg.UseProxy, "use-proxy", getenvBool("QIQIAO_USE_PROXY", false), "use system HTTP proxy instead of direct intranet connection")
	flag.IntVar(&cfg.TimeoutSeconds, "timeout", getenvInt("QIQIAO_TIMEOUT_SECONDS", cfg.TimeoutSeconds), "HTTP timeout seconds")
	flag.StringVar(&variablesJSON, "variables-json", "", "JSON object for crud-smoke create/update variables")
	flag.IntVar(&pageSize, "page-size", 20, "schema/query page size")
	flag.BoolVar(&deleteAfterSmoke, "delete-after", true, "delete crud-smoke record after create/update")
	flag.StringVar(&uploadFilePath, "upload-file", "", "optional local file for full-smoke file upload test")
	flag.Parse()

	loaded, err := LoadConfig(configPath, qqkfPath)
	if err != nil {
		fatal(err)
	}
	cfg = MergeConfig(loaded, cfg)
	cfg.BaseURL = NormalizeBaseURL(cfg.BaseURL)
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 20
	}

	client := NewQiqiaoClient(cfg)
	ctx := context.Background()

	switch mode {
	case "serve":
		if err := validateServeConfig(cfg); err != nil {
			fatal(err)
		}
		log.Printf("数智彩虹测试网页已启动: http://%s", listen)
		log.Printf("OpenAPI: %s, app=%s, form=%s, proxy=%v", cfg.BaseURL, cfg.ApplicationID, cfg.FormModelID, cfg.UseProxy)
		fatal(http.ListenAndServe(listen, NewServer(client)))
	case "probe":
		fatal(runProbe(ctx, client, pageSize))
	case "schema":
		fatal(runSchema(ctx, client))
	case "crud-smoke":
		fatal(runCRUDSmoke(ctx, client, variablesJSON, deleteAfterSmoke))
	case "full-smoke":
		fatal(runFullSmoke(ctx, client, variablesJSON, uploadFilePath, deleteAfterSmoke, pageSize))
	case "design-probe":
		fatal(runDesignProbe(ctx, client))
	default:
		fatal(fmt.Errorf("unknown mode %q", mode))
	}
}

func ParseQQKF(r io.Reader) (Config, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return Config{}, err
	}
	text := string(body)
	cfg := Config{TimeoutSeconds: 20}

	cfg.CorpID = firstGroup(text, `(?m)(?:CorpID|CropID|企业I[Dd]|企业Id)[：:]\s*([^\s]+)`)
	cfg.Secret = firstGroup(text, `(?m)Secret[：:]\s*([0-9a-fA-F]{32})`)
	cfg.Account = firstGroup(text, `(?m)Account[：:]\s*([^\s]+)`)

	idRe := regexp.MustCompile(`\b[0-9a-fA-F]{24}\b`)
	ids := idRe.FindAllString(text, -1)
	if len(ids) > 0 {
		cfg.ApplicationID = ids[0]
	}
	if len(ids) > 1 {
		cfg.FormModelID = ids[1]
	}

	if cfg.CorpID == "" || cfg.Secret == "" || cfg.Account == "" {
		return cfg, errors.New("qqkf.txt is missing CorpID/CropID, Secret, or Account")
	}
	return cfg, nil
}

func LoadConfig(configPath, qqkfPath string) (Config, error) {
	cfg := Config{TimeoutSeconds: 20}
	if qqkfPath != "" {
		f, err := os.Open(qqkfPath)
		if err != nil {
			return cfg, err
		}
		defer f.Close()
		fromQQKF, err := ParseQQKF(f)
		if err != nil {
			return cfg, err
		}
		cfg = MergeConfig(cfg, fromQQKF)
	}
	if configPath != "" {
		body, err := os.ReadFile(configPath)
		if err != nil {
			return cfg, err
		}
		var fromFile Config
		if err := json.Unmarshal(body, &fromFile); err != nil {
			return cfg, err
		}
		cfg = MergeConfig(cfg, fromFile)
	}
	return cfg, nil
}

func MergeConfig(base, override Config) Config {
	out := base
	if override.BaseURL != "" {
		out.BaseURL = override.BaseURL
	}
	if override.CorpID != "" {
		out.CorpID = override.CorpID
	}
	if override.Secret != "" {
		out.Secret = override.Secret
	}
	if override.Account != "" {
		out.Account = override.Account
	}
	if override.ApplicationID != "" {
		out.ApplicationID = override.ApplicationID
	}
	if override.FormModelID != "" {
		out.FormModelID = override.FormModelID
	}
	if override.InsecureSkipVerify {
		out.InsecureSkipVerify = true
	}
	if override.UseProxy {
		out.UseProxy = true
	}
	if override.TimeoutSeconds != 0 {
		out.TimeoutSeconds = override.TimeoutSeconds
	}
	return out
}

func NormalizeBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

func BuildSaveRecordPayload(variables map[string]any, id, loginUserID string) map[string]any {
	return map[string]any{
		"variables":   cleanVariables(variables),
		"id":          id,
		"loginUserId": loginUserID,
	}
}

func BuildUpdateRecordPayload(variables map[string]any, id, loginUserID string, currentVersion int) map[string]any {
	return map[string]any{
		"variables":   cleanVariables(variables),
		"id":          id,
		"version":     currentVersion,
		"loginUserId": loginUserID,
	}
}

func NewQiqiaoClient(cfg Config) *QiqiaoClient {
	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}, //nolint:gosec
		ForceAttemptHTTP2: false,
		TLSNextProto:      map[string]func(string, *tls.Conn) http.RoundTripper{},
	}
	if cfg.UseProxy {
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &QiqiaoClient{
		cfg: cfg,
		client: &http.Client{
			Timeout:   time.Duration(cfg.TimeoutSeconds) * time.Second,
			Transport: transport,
		},
	}
}

func (c *QiqiaoClient) Token(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.token != "" && time.Since(c.tokenTime) < 50*time.Minute {
		token := c.token
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	if err := validateAuthConfig(c.cfg); err != nil {
		return "", err
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	randomValue := strconv.Itoa(time.Now().Nanosecond()%100 + 1)
	common := url.Values{
		"timestamp": {timestamp},
		"random":    {randomValue},
		"corpId":    {c.cfg.CorpID},
		"secret":    {c.cfg.Secret},
		"account":   {c.cfg.Account},
	}

	var access APIEnvelope
	if err := c.getJSON(ctx, c.cfg.BaseURL+"/securities/access_key?"+common.Encode(), nil, &access); err != nil {
		return "", err
	}
	if !envelopeOK(access) {
		return "", fmt.Errorf("access_key failed: code=%v msg=%s", access.Code, access.Msg)
	}
	var accessKey string
	if err := json.Unmarshal(access.Data, &accessKey); err != nil || accessKey == "" {
		return "", fmt.Errorf("access_key returned invalid data")
	}

	tokenQuery := common
	tokenQuery.Set("accessKey", accessKey)
	var tokenResp APIEnvelope
	if err := c.getJSON(ctx, c.cfg.BaseURL+"/securities/qiqiao_token?"+tokenQuery.Encode(), nil, &tokenResp); err != nil {
		return "", err
	}
	if !envelopeOK(tokenResp) {
		return "", fmt.Errorf("qiqiao_token failed: code=%v msg=%s", tokenResp.Code, tokenResp.Msg)
	}
	var token string
	if err := json.Unmarshal(tokenResp.Data, &token); err != nil || token == "" {
		return "", fmt.Errorf("qiqiao_token returned invalid data")
	}

	c.mu.Lock()
	c.token = token
	c.tokenTime = time.Now()
	c.mu.Unlock()
	return token, nil
}

func (c *QiqiaoClient) ListFormModels(ctx context.Context) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodGet, fmt.Sprintf("/open/applications/%s/form_models", c.cfg.ApplicationID), nil)
}

func (c *QiqiaoClient) FormModel(ctx context.Context) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodGet, fmt.Sprintf("/open/applications/%s/form_models/%s", c.cfg.ApplicationID, c.cfg.FormModelID), nil)
}

func (c *QiqiaoClient) QueryRecords(ctx context.Context, page, pageSize int, filters any) (APIEnvelope, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	path := fmt.Sprintf("/open/applications/%s/forms/%s/query?page=%d&pageSize=%d", c.cfg.ApplicationID, c.cfg.FormModelID, page, pageSize)
	if filters == nil {
		filters = []any{}
	}
	return c.openJSON(ctx, http.MethodPost, path, filters)
}

func (c *QiqiaoClient) PageRecords(ctx context.Context, page, pageSize int) (APIEnvelope, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	path := fmt.Sprintf("/open/applications/%s/forms/%s?page=%d&pageSize=%d", c.cfg.ApplicationID, c.cfg.FormModelID, page, pageSize)
	return c.openJSON(ctx, http.MethodGet, path, nil)
}

func (c *QiqiaoClient) CurrentUserID(ctx context.Context) (string, APIEnvelope, error) {
	path := fmt.Sprintf("/open/users/account?account=%s", url.QueryEscape(c.cfg.Account))
	env, err := c.openJSON(ctx, http.MethodGet, path, nil)
	if err != nil || !envelopeOK(env) {
		return "", env, err
	}
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return "", env, err
	}
	id, _ := data["id"].(string)
	if id == "" {
		return "", env, errors.New("current user response has no id")
	}
	return id, env, nil
}

func (c *QiqiaoClient) CreateRecord(ctx context.Context, req RecordRequest) (APIEnvelope, error) {
	if req.ID == "" {
		req.ID = randomID32()
	}
	if req.LoginUserID == "" {
		userID, _, err := c.CurrentUserID(ctx)
		if err != nil {
			return APIEnvelope{}, err
		}
		req.LoginUserID = userID
	}
	payload := BuildSaveRecordPayload(req.Variables, req.ID, req.LoginUserID)
	return c.openJSON(ctx, http.MethodPost, fmt.Sprintf("/open/applications/%s/forms/%s", c.cfg.ApplicationID, c.cfg.FormModelID), payload)
}

func (c *QiqiaoClient) UpdateRecord(ctx context.Context, id string, req RecordRequest) (APIEnvelope, error) {
	if id == "" {
		id = req.ID
	}
	if id == "" {
		return APIEnvelope{}, errors.New("record id is required")
	}
	if req.LoginUserID == "" {
		userID, _, err := c.CurrentUserID(ctx)
		if err != nil {
			return APIEnvelope{}, err
		}
		req.LoginUserID = userID
	}
	payload := BuildUpdateRecordPayload(req.Variables, id, req.LoginUserID, req.Version)
	return c.openJSON(ctx, http.MethodPut, fmt.Sprintf("/open/applications/%s/forms/%s", c.cfg.ApplicationID, c.cfg.FormModelID), payload)
}

func (c *QiqiaoClient) DeleteRecord(ctx context.Context, id string) (APIEnvelope, error) {
	if id == "" {
		return APIEnvelope{}, errors.New("record id is required")
	}
	return c.openJSON(ctx, http.MethodDelete, fmt.Sprintf("/open/applications/%s/forms/%s/%s", c.cfg.ApplicationID, c.cfg.FormModelID, url.PathEscape(id)), nil)
}

func (c *QiqiaoClient) GetRecord(ctx context.Context, id string) (APIEnvelope, error) {
	if id == "" {
		return APIEnvelope{}, errors.New("record id is required")
	}
	return c.openJSON(ctx, http.MethodGet, fmt.Sprintf("/open/applications/%s/forms/%s/%s", c.cfg.ApplicationID, c.cfg.FormModelID, url.PathEscape(id)), nil)
}

func (c *QiqiaoClient) BatchSaveRecords(ctx context.Context, records []map[string]any) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodPost, fmt.Sprintf("/open/applications/%s/forms/%s/batch_save", c.cfg.ApplicationID, c.cfg.FormModelID), records)
}

func (c *QiqiaoClient) BatchUpdateRecords(ctx context.Context, records []map[string]any) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodPost, fmt.Sprintf("/open/applications/%s/forms/%s/batch_update", c.cfg.ApplicationID, c.cfg.FormModelID), records)
}

func (c *QiqiaoClient) BatchUpdateByCondition(ctx context.Context, payload map[string]any) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodPut, fmt.Sprintf("/open/applications/%s/forms/batch_update_by_condition", c.cfg.ApplicationID), payload)
}

func (c *QiqiaoClient) BatchDeleteRecords(ctx context.Context, ids []string) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodPost, fmt.Sprintf("/open/applications/%s/forms/%s/batch_delete", c.cfg.ApplicationID, c.cfg.FormModelID), ids)
}

func (c *QiqiaoClient) Applications(ctx context.Context) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodGet, "/open/applications", nil)
}

func (c *QiqiaoClient) RootDepartment(ctx context.Context) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodGet, "/open/departments/root", nil)
}

func (c *QiqiaoClient) UploadFile(ctx context.Context, path string, fieldType string) (APIEnvelope, error) {
	if err := validateOpenAPIConfig(c.cfg); err != nil {
		return APIEnvelope{}, err
	}
	if path == "" {
		return APIEnvelope{}, errors.New("upload file path is required")
	}
	token, err := c.Token(ctx)
	if err != nil {
		return APIEnvelope{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return APIEnvelope{}, err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("files", filepath.Base(path))
	if err != nil {
		return APIEnvelope{}, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return APIEnvelope{}, err
	}
	if err := writer.Close(); err != nil {
		return APIEnvelope{}, err
	}

	if fieldType == "" {
		fieldType = "FILEUPLOAD"
	}
	rawURL := fmt.Sprintf("%s/open/file_upload/applications/%s/form_models/%s?fieldType=%s", c.cfg.BaseURL, c.cfg.ApplicationID, c.cfg.FormModelID, url.QueryEscape(fieldType))
	headers := map[string]string{
		"Content-Type":  writer.FormDataContentType(),
		"Accept":        "application/json",
		"X-Auth0-Token": token,
	}
	var env APIEnvelope
	err = c.doRaw(ctx, http.MethodPost, rawURL, headers, &buf, &env)
	return env, err
}

func (c *QiqiaoClient) WorkflowDefinitions(ctx context.Context) (APIEnvelope, error) {
	return c.openJSON(ctx, http.MethodGet, fmt.Sprintf("/open/console/applications/%s/definitions", c.cfg.ApplicationID), nil)
}

func (c *QiqiaoClient) openJSON(ctx context.Context, method, path string, payload any) (APIEnvelope, error) {
	if err := validateOpenAPIConfig(c.cfg); err != nil {
		return APIEnvelope{}, err
	}
	token, err := c.Token(ctx)
	if err != nil {
		return APIEnvelope{}, err
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"X-Auth0-Token": token,
	}
	var env APIEnvelope
	err = c.doJSON(ctx, method, c.cfg.BaseURL+path, headers, payload, &env)
	return env, err
}

func (c *QiqiaoClient) doRaw(ctx context.Context, method, rawURL string, headers map[string]string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 qiqiao-openapi-tool/1.0")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil && len(respBody) == 0 {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s returned HTTP %d: %s", method, redactURL(rawURL), resp.StatusCode, trimForLog(string(respBody)))
	}
	return decodeJSONBody(respBody, err, out, redactURL(rawURL))
}

func (c *QiqiaoClient) getJSON(ctx context.Context, rawURL string, headers map[string]string, out any) error {
	return c.doJSON(ctx, http.MethodGet, rawURL, headers, nil, out)
}

func (c *QiqiaoClient) doJSON(ctx context.Context, method, rawURL string, headers map[string]string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 qiqiao-openapi-tool/1.0")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil && len(respBody) == 0 {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s returned HTTP %d: %s", method, redactURL(rawURL), resp.StatusCode, trimForLog(string(respBody)))
	}
	return decodeJSONBody(respBody, err, out, redactURL(rawURL))
}

func decodeJSONBody(respBody []byte, readErr error, out any, source string) error {
	if len(respBody) == 0 {
		return readErr
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		if readErr != nil {
			return fmt.Errorf("read from %s ended with %v and JSON decode failed: %w; body=%s", source, readErr, err, trimForLog(string(respBody)))
		}
		return fmt.Errorf("decode JSON from %s failed: %w; body=%s", source, err, trimForLog(string(respBody)))
	}
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return fmt.Errorf("read from %s failed after JSON decode: %w", source, readErr)
	}
	return nil
}

func NewServer(client *QiqiaoClient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(uiHTML))
	})
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"baseUrl":       client.cfg.BaseURL,
			"applicationId": client.cfg.ApplicationID,
			"formModelId":   client.cfg.FormModelID,
			"account":       client.cfg.Account,
			"useProxy":      client.cfg.UseProxy,
		})
	})
	mux.HandleFunc("/api/schema", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		models, err := client.ListFormModels(ctx)
		if err != nil {
			writeError(w, http.StatusBadGateway, "获取表单列表失败", err)
			return
		}
		detail, err := client.FormModel(ctx)
		if err != nil {
			writeError(w, http.StatusBadGateway, "获取表单组件失败", err)
			return
		}
		writeJSON(w, map[string]any{"models": models, "detail": detail})
	})
	mux.HandleFunc("/api/probe", func(w http.ResponseWriter, r *http.Request) {
		result, err := probeResult(r.Context(), client, 5)
		if err != nil {
			writeError(w, http.StatusBadGateway, "探测失败", err)
			return
		}
		writeJSON(w, result)
	})
	mux.HandleFunc("/api/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			page := queryInt(r, "page", 1)
			pageSize := queryInt(r, "pageSize", 20)
			env, err := client.QueryRecords(r.Context(), page, pageSize, nil)
			if err != nil {
				writeError(w, http.StatusBadGateway, "查询记录失败", err)
				return
			}
			writeJSON(w, env)
		case http.MethodPost:
			var req RecordRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "请求 JSON 无效", err)
				return
			}
			env, err := client.CreateRecord(r.Context(), req)
			if err != nil {
				writeError(w, http.StatusBadGateway, "新增记录失败", err)
				return
			}
			writeJSON(w, env)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/records/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/records/")
		switch r.Method {
		case http.MethodPut:
			var req RecordRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "请求 JSON 无效", err)
				return
			}
			env, err := client.UpdateRecord(r.Context(), id, req)
			if err != nil {
				writeError(w, http.StatusBadGateway, "修改记录失败", err)
				return
			}
			writeJSON(w, env)
		case http.MethodDelete:
			env, err := client.DeleteRecord(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusBadGateway, "删除记录失败", err)
				return
			}
			writeJSON(w, env)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func runProbe(ctx context.Context, client *QiqiaoClient, pageSize int) error {
	result, err := probeResult(ctx, client, pageSize)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func probeResult(ctx context.Context, client *QiqiaoClient, pageSize int) (map[string]any, error) {
	token, err := client.Token(ctx)
	if err != nil {
		return nil, err
	}
	userID, userEnv, userErr := client.CurrentUserID(ctx)
	models, modelsErr := client.ListFormModels(ctx)
	detail, detailErr := client.FormModel(ctx)
	records, recordsErr := client.QueryRecords(ctx, 1, pageSize, nil)

	result := map[string]any{
		"ok":            userErr == nil && modelsErr == nil && detailErr == nil && recordsErr == nil,
		"token":         redacted(token),
		"baseUrl":       client.cfg.BaseURL,
		"applicationId": client.cfg.ApplicationID,
		"formModelId":   client.cfg.FormModelID,
		"useProxy":      client.cfg.UseProxy,
		"userId":        redacted(userID),
		"user":          summarizeEnvelope(userEnv, userErr),
		"formModels":    summarizeEnvelope(models, modelsErr),
		"formDetail":    summarizeEnvelope(detail, detailErr),
		"records":       summarizeEnvelope(records, recordsErr),
	}
	return result, nil
}

func runSchema(ctx context.Context, client *QiqiaoClient) error {
	models, err := client.ListFormModels(ctx)
	if err != nil {
		return err
	}
	detail, err := client.FormModel(ctx)
	if err != nil {
		return err
	}
	return printJSON(map[string]any{"formModels": models, "formDetail": detail})
}

func runCRUDSmoke(ctx context.Context, client *QiqiaoClient, variablesJSON string, deleteAfter bool) error {
	if variablesJSON == "" {
		return errors.New("crud-smoke requires --variables-json, for example '{\"测试文本\":\"UOS smoke\"}'")
	}
	var variables map[string]any
	if err := json.Unmarshal([]byte(variablesJSON), &variables); err != nil {
		return err
	}
	id := randomID32()
	create, err := client.CreateRecord(ctx, RecordRequest{ID: id, Variables: variables})
	if err != nil {
		return err
	}
	updateVars := mutateVariablesForSmoke(variables)
	version, ok := versionFromEnvelope(create)
	if !ok {
		version = 0
	}
	update, err := client.UpdateRecord(ctx, id, RecordRequest{Variables: updateVars, Version: version})
	if err != nil {
		return err
	}
	result := map[string]any{
		"id":     id,
		"create": summarizeEnvelope(create, nil),
		"update": summarizeEnvelope(update, nil),
	}
	if deleteAfter {
		deleted, err := client.DeleteRecord(ctx, id)
		result["delete"] = summarizeEnvelope(deleted, err)
		if err != nil {
			_ = printJSON(result)
			return err
		}
	}
	return printJSON(result)
}

func runFullSmoke(ctx context.Context, client *QiqiaoClient, variablesJSON, uploadFilePath string, deleteAfter bool, pageSize int) error {
	result := map[string]any{
		"baseUrl":       client.cfg.BaseURL,
		"applicationId": client.cfg.ApplicationID,
		"formModelId":   client.cfg.FormModelID,
		"startedAt":     time.Now().Format(time.RFC3339),
		"note":          formDesignProbeNote(),
	}
	steps := []TestStep{}
	add := func(step TestStep) {
		steps = append(steps, step)
	}
	addCall := func(name string, env APIEnvelope, err error) {
		add(stepFromEnvelope(name, env, err))
	}
	runStep := func(name string, fn func() (APIEnvelope, error)) {
		env, err := fn()
		addCall(name, env, err)
	}

	token, err := client.Token(ctx)
	add(TestStep{Name: "auth.token", OK: err == nil && token != "", Error: errString(err), Message: redacted(token)})
	if err != nil {
		result["steps"] = steps
		result["ok"] = false
		return printJSON(result)
	}

	userID, userEnv, err := client.CurrentUserID(ctx)
	add(stepFromEnvelope("users.current_by_account", userEnv, err))
	runStep("applications.list", func() (APIEnvelope, error) { return client.Applications(ctx) })
	runStep("departments.root", func() (APIEnvelope, error) { return client.RootDepartment(ctx) })
	models, err := client.ListFormModels(ctx)
	add(stepFromEnvelope("form_models.list", models, err))
	detail, err := client.FormModel(ctx)
	add(stepFromEnvelope("form_models.detail", detail, err))
	fields := fieldsFromEnvelope(detail)
	result["fields"] = fields

	runStep("forms.page_get", func() (APIEnvelope, error) { return client.PageRecords(ctx, 1, pageSize) })
	runStep("forms.query_post", func() (APIEnvelope, error) { return client.QueryRecords(ctx, 1, pageSize, []any{}) })

	var variables map[string]any
	if variablesJSON != "" {
		if err := json.Unmarshal([]byte(variablesJSON), &variables); err != nil {
			add(TestStep{Name: "variables.parse", OK: false, Error: err.Error()})
		}
	}
	if len(variables) == 0 {
		variables = BuildSmokeVariables(fields, "full-smoke")
	}
	result["variablesUsed"] = variables

	createdIDs := []string{}
	id := randomID32()
	create, err := client.CreateRecord(ctx, RecordRequest{ID: id, LoginUserID: userID, Variables: variables})
	add(stepFromEnvelope("forms.create_one", create, err))
	if err == nil && envelopeOK(create) {
		createdIDs = append(createdIDs, id)
	}
	version, _ := versionFromEnvelope(create)
	runStep("forms.get_by_id", func() (APIEnvelope, error) { return client.GetRecord(ctx, id) })
	filter := []map[string]any{{"fieldName": "id", "logic": "eq", "value": id}}
	runStep("forms.query_filter_id", func() (APIEnvelope, error) { return client.QueryRecords(ctx, 1, 5, filter) })
	runStep("forms.update_one", func() (APIEnvelope, error) {
		return client.UpdateRecord(ctx, id, RecordRequest{LoginUserID: userID, Version: version, Variables: mutateVariablesForSmoke(variables)})
	})
	conditionPayload := map[string]any{
		"formModelId":  client.cfg.FormModelID,
		"filters":      []map[string]any{{"fieldName": "id", "logic": "eq", "value": id}},
		"variables":    mutateVariablesForSmoke(variables),
		"loginUserId":  userID,
		"probeComment": "qqkf-listed path; current deployment may require a private payload shape",
	}
	runStep("forms.batch_update_by_condition_probe", func() (APIEnvelope, error) {
		return client.BatchUpdateByCondition(ctx, conditionPayload)
	})

	batchID1, batchID2 := randomID32(), randomID32()
	batchRecords := []map[string]any{
		BuildSaveRecordPayload(BuildSmokeVariables(fields, "batch-1"), batchID1, userID),
		BuildSaveRecordPayload(BuildSmokeVariables(fields, "batch-2"), batchID2, userID),
	}
	batchCreate, err := client.BatchSaveRecords(ctx, batchRecords)
	add(stepFromEnvelope("forms.batch_save", batchCreate, err))
	if err == nil && envelopeOK(batchCreate) {
		createdIDs = append(createdIDs, batchID1, batchID2)
	}
	batchUpdatePayload := buildBatchUpdatePayload(batchCreate, fields, userID)
	if len(batchUpdatePayload) > 0 {
		runStep("forms.batch_update", func() (APIEnvelope, error) { return client.BatchUpdateRecords(ctx, batchUpdatePayload) })
	}

	if uploadFilePath != "" {
		upload, err := client.UploadFile(ctx, uploadFilePath, "FILEUPLOAD")
		add(stepFromEnvelope("files.upload", upload, err))
	}

	runStep("workflow_design.definitions_list", func() (APIEnvelope, error) { return client.WorkflowDefinitions(ctx) })
	add(TestStep{Name: "form_design.create_form_or_fields", OK: false, Message: formDesignProbeNote()})

	if deleteAfter && len(createdIDs) > 0 {
		runStep("forms.batch_delete_cleanup", func() (APIEnvelope, error) { return client.BatchDeleteRecords(ctx, createdIDs) })
	}

	result["steps"] = steps
	result["ok"] = fullSmokeOK(steps)
	result["finishedAt"] = time.Now().Format(time.RFC3339)
	return printJSON(result)
}

func runDesignProbe(ctx context.Context, client *QiqiaoClient) error {
	env, err := client.WorkflowDefinitions(ctx)
	result := map[string]any{
		"note": "This endpoint is workflow design definitions, not confirmed form-model design creation.",
		"path": fmt.Sprintf("/open/console/applications/%s/definitions", client.cfg.ApplicationID),
		"data": summarizeEnvelope(env, err),
	}
	if err != nil {
		_ = printJSON(result)
		return err
	}
	return printJSON(result)
}

func BuildSmokeVariables(fields []FormField, marker string) map[string]any {
	vars := map[string]any{}
	now := time.Now()
	for _, field := range fields {
		name := field.Name
		if name == "" {
			continue
		}
		switch strings.ToLower(field.Type) {
		case "textbox", "text", "input":
			vars[name] = "api-" + marker
		case "textarea":
			vars[name] = "api-" + marker + "\nmultiline"
		case "time":
			vars[name] = now.Format("15:04")
		case "date", "datetime", "datetimepicker":
			vars[name] = now.UnixMilli()
		}
	}
	if len(vars) == 0 {
		vars["单行文本1"] = "api-" + marker
	}
	return vars
}

func buildBatchUpdatePayload(batchCreate APIEnvelope, fields []FormField, loginUserID string) []map[string]any {
	var data []map[string]any
	if err := json.Unmarshal(batchCreate.Data, &data); err != nil {
		return nil
	}
	payload := make([]map[string]any, 0, len(data))
	for i, item := range data {
		id, _ := item["id"].(string)
		if id == "" {
			continue
		}
		version := 0
		switch v := item["version"].(type) {
		case float64:
			version = int(v)
		case int:
			version = v
		}
		payload = append(payload, BuildUpdateRecordPayload(BuildSmokeVariables(fields, fmt.Sprintf("batch-updated-%d", i+1)), id, loginUserID, version+1))
	}
	return payload
}

func fieldsFromEnvelope(env APIEnvelope) []FormField {
	var raw []map[string]any
	if err := json.Unmarshal(env.Data, &raw); err != nil {
		return nil
	}
	fields := make([]FormField, 0, len(raw))
	for _, item := range raw {
		fields = append(fields, FormField{
			Name: firstString(item, "name", "fieldName", "label", "title", "componentName"),
			Type: firstString(item, "type", "fieldType", "componentType", "controlType"),
			ID:   firstString(item, "id", "fieldId"),
		})
	}
	return fields
}

func firstString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := item[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func stepFromEnvelope(name string, env APIEnvelope, err error) TestStep {
	step := TestStep{Name: name}
	if err != nil {
		step.OK = false
		step.Error = err.Error()
		return step
	}
	step.OK = envelopeOK(env)
	step.Code = env.Code
	step.Message = env.Msg
	var data any
	if len(env.Data) > 0 && json.Unmarshal(env.Data, &data) == nil {
		step.Shape = dataShape(data)
	}
	return step
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func fullSmokeOK(steps []TestStep) bool {
	for _, step := range steps {
		if strings.HasPrefix(step.Name, "form_design.") || step.Name == "form_models.list" || step.Name == "applications.list" || step.Name == "files.upload" || step.Name == "forms.batch_update_by_condition_probe" {
			continue
		}
		if !step.OK {
			return false
		}
	}
	return true
}

func formDesignProbeNote() string {
	return "non-destructive probe only: documented OpenAPI can read form models/components and CRUD form records, but does not create form models or fields without a verified management endpoint"
}

func cleanVariables(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		if k == "id" || k == "loginUserId" || k == "version" {
			continue
		}
		out[k] = v
	}
	return out
}

func mutateVariablesForSmoke(in map[string]any) map[string]any {
	out := map[string]any{}
	changed := false
	for k, v := range in {
		if text, ok := v.(string); ok && !changed {
			out[k] = text + "-updated"
			changed = true
			continue
		}
		out[k] = v
	}
	return out
}

func firstGroup(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func envelopeOK(env APIEnvelope) bool {
	switch v := env.Code.(type) {
	case float64:
		return v == 0
	case int:
		return v == 0
	case string:
		return v == "0"
	default:
		return false
	}
}

func validateAuthConfig(cfg Config) error {
	missing := []string{}
	if cfg.BaseURL == "" {
		missing = append(missing, "base-url")
	}
	if cfg.CorpID == "" {
		missing = append(missing, "corp-id")
	}
	if cfg.Secret == "" {
		missing = append(missing, "secret")
	}
	if cfg.Account == "" {
		missing = append(missing, "account")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing auth config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func validateOpenAPIConfig(cfg Config) error {
	if err := validateAuthConfig(cfg); err != nil {
		return err
	}
	missing := []string{}
	if cfg.ApplicationID == "" {
		missing = append(missing, "app-id")
	}
	if cfg.FormModelID == "" {
		missing = append(missing, "form-id")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing openapi config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func validateServeConfig(cfg Config) error {
	return validateOpenAPIConfig(cfg)
}

func randomID32() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}

func summarizeEnvelope(env APIEnvelope, err error) map[string]any {
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error()}
	}
	out := map[string]any{
		"ok":   envelopeOK(env),
		"code": env.Code,
		"msg":  env.Msg,
	}
	var data any
	if len(env.Data) > 0 && json.Unmarshal(env.Data, &data) == nil {
		out["shape"] = dataShape(data)
	}
	return out
}

func versionFromEnvelope(env APIEnvelope) (int, bool) {
	if len(env.Data) == 0 {
		return 0, false
	}
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return 0, false
	}
	switch v := data["version"].(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		n, err := strconv.Atoi(v)
		return n, err == nil
	default:
		return 0, false
	}
}

func dataShape(v any) any {
	switch x := v.(type) {
	case []any:
		return map[string]any{"type": "array", "length": len(x)}
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
			if len(keys) >= 20 {
				break
			}
		}
		return map[string]any{"type": "object", "keys": keys}
	default:
		return fmt.Sprintf("%T", v)
	}
}

func queryInt(r *http.Request, name string, fallback int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || v <= 0 || math.IsNaN(float64(v)) {
		return fallback
	}
	return v
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	w.WriteHeader(status)
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	writeJSON(w, ErrorResponse{OK: false, Message: message, Detail: detail})
}

func printJSON(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getenv(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func getenvBool(name string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func getenvInt(name string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		return fallback
	}
	return v
}

func redacted(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 10 {
		return "***"
	}
	return value[:6] + "***" + value[len(value)-4:]
}

func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for _, key := range []string{"secret", "accessKey", "X-Auth0-Token"} {
		if q.Has(key) {
			q.Set(key, "***")
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func trimForLog(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 500 {
		return s[:500] + "..."
	}
	return s
}

const uiHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>数智彩虹 OpenAPI 测试台</title>
  <style>
    :root { color-scheme: light; --ink:#1b2430; --muted:#667085; --line:#d7dde7; --brand:#126e82; --ok:#0f766e; --bad:#b42318; --bg:#f5f7fa; }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: var(--ink); background: var(--bg); }
    header { padding: 18px 24px; border-bottom: 1px solid var(--line); background: #fff; display: flex; align-items: center; justify-content: space-between; gap: 16px; }
    h1 { margin: 0; font-size: 20px; letter-spacing: 0; }
    main { max-width: 1180px; margin: 0 auto; padding: 20px; display: grid; grid-template-columns: 320px minmax(0, 1fr); gap: 16px; }
    section, aside { background: #fff; border: 1px solid var(--line); border-radius: 8px; }
    aside { padding: 16px; align-self: start; }
    section { padding: 16px; min-width: 0; }
    label { display: block; font-size: 13px; color: var(--muted); margin: 12px 0 6px; }
    input, textarea { width: 100%; border: 1px solid var(--line); border-radius: 6px; padding: 10px; font: inherit; background: #fff; }
    textarea { min-height: 150px; resize: vertical; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 13px; }
    button { border: 0; border-radius: 6px; background: var(--brand); color: #fff; padding: 9px 12px; font: inherit; cursor: pointer; }
    button.secondary { background: #475467; }
    button.danger { background: var(--bad); }
    .row { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
    .status { font-size: 13px; color: var(--muted); overflow-wrap: anywhere; }
    .pill { display: inline-flex; align-items: center; min-height: 24px; padding: 2px 8px; border-radius: 999px; background: #eef4ff; color: #194185; font-size: 12px; }
    pre { margin: 12px 0 0; white-space: pre-wrap; overflow-wrap: anywhere; background: #0f172a; color: #d9e6ff; padding: 12px; border-radius: 8px; min-height: 160px; max-height: 460px; overflow: auto; }
    table { width: 100%; border-collapse: collapse; margin-top: 12px; font-size: 13px; }
    th, td { border-bottom: 1px solid var(--line); text-align: left; vertical-align: top; padding: 8px; }
    th { color: var(--muted); font-weight: 600; }
    td code { white-space: nowrap; }
    @media (max-width: 860px) { main { grid-template-columns: 1fr; padding: 12px; } header { align-items: flex-start; flex-direction: column; } }
  </style>
</head>
<body>
  <header>
    <h1>数智彩虹 OpenAPI 测试台</h1>
    <div id="config" class="status"></div>
  </header>
  <main>
    <aside>
      <div class="row">
        <button id="probe">连通探测</button>
        <button id="schema" class="secondary">读取表单结构</button>
      </div>
      <label>新增/修改变量 JSON</label>
      <textarea id="variables">{"测试文本":"UOS网页测试"}</textarea>
      <label>记录 ID</label>
      <input id="recordId" placeholder="新增时可留空，修改/删除时填写">
      <label>当前 version</label>
      <input id="version" type="number" min="0" value="0">
      <div class="row" style="margin-top:12px">
        <button id="create">新增</button>
        <button id="update" class="secondary">修改</button>
        <button id="delete" class="danger">删除</button>
      </div>
      <div class="row" style="margin-top:12px">
        <button id="query" class="secondary">查询前 20 条</button>
      </div>
      <p class="status">字段名必须与七巧表单组件名称一致。单选、多选、日期等字段请按目标表单实际值格式填写。</p>
    </aside>
    <section>
      <div class="row"><span class="pill" id="lastAction">待操作</span><span class="status" id="message"></span></div>
      <pre id="output"></pre>
      <table id="records"></table>
    </section>
  </main>
  <script>
    const out = document.getElementById("output");
    const msg = document.getElementById("message");
    const lastAction = document.getElementById("lastAction");
    const table = document.getElementById("records");
    function show(action, value) {
      lastAction.textContent = action;
      out.textContent = JSON.stringify(value, null, 2);
      msg.textContent = value && value.ok === false ? value.message || "失败" : "完成";
    }
    async function api(action, path, options) {
      try {
        const res = await fetch(path, options);
        const data = await res.json();
        show(action, data);
        if (!res.ok) throw new Error(data.detail || data.message || res.statusText);
        return data;
      } catch (error) {
        show(action, { ok: false, message: error.message });
      }
    }
    function variables() {
      const text = document.getElementById("variables").value.trim();
      return text ? JSON.parse(text) : {};
    }
    function renderRecords(env) {
      table.innerHTML = "";
      const data = env && env.data;
      const list = data && (data.list || data.rows);
      if (!Array.isArray(list)) return;
      table.innerHTML = "<thead><tr><th>ID</th><th>version</th><th>variables</th></tr></thead><tbody>" +
        list.map(item => "<tr><td><code>" + (item.id || "") + "</code></td><td>" + (item.version ?? "") + "</td><td><code>" +
        JSON.stringify(item.variables || item).replace(/[<>&]/g, c => ({ "<":"&lt;", ">":"&gt;", "&":"&amp;" }[c])) + "</code></td></tr>").join("") +
        "</tbody>";
    }
    document.getElementById("probe").onclick = () => api("连通探测", "/api/probe");
    document.getElementById("schema").onclick = () => api("读取表单结构", "/api/schema");
    document.getElementById("query").onclick = async () => renderRecords(await api("查询记录", "/api/records?page=1&pageSize=20"));
    document.getElementById("create").onclick = () => api("新增", "/api/records", { method:"POST", headers:{ "Content-Type":"application/json" }, body: JSON.stringify({ variables: variables(), id: document.getElementById("recordId").value }) });
    document.getElementById("update").onclick = () => {
      const id = document.getElementById("recordId").value.trim();
      api("修改", "/api/records/" + encodeURIComponent(id), { method:"PUT", headers:{ "Content-Type":"application/json" }, body: JSON.stringify({ variables: variables(), version: Number(document.getElementById("version").value || 0) }) });
    };
    document.getElementById("delete").onclick = () => {
      const id = document.getElementById("recordId").value.trim();
      api("删除", "/api/records/" + encodeURIComponent(id), { method:"DELETE" });
    };
    fetch("/api/config").then(r => r.json()).then(c => {
      document.getElementById("config").textContent = c.baseUrl + " / app " + c.applicationId + " / form " + c.formModelId;
    });
  </script>
</body>
</html>`
