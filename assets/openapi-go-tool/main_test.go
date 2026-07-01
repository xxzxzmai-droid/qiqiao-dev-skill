package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestParseQQKFUsesCredentialLinesAndSkipsExampleTokens(t *testing.T) {
	input := `
CropID：corp-test-001
Secret：11111111111111111111111111111111
Account：tester@example.com
"data": "cccccccccccccccccccccccccccccccc"
"X-Auth0-Token":"dddddddddddddddddddddddddddddddd"
demo-app
aaaaaaaaaaaaaaaaaaaaaaaa
未命名表单
bbbbbbbbbbbbbbbbbbbbbbbb
`

	cfg, err := ParseQQKF(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseQQKF returned error: %v", err)
	}

	if cfg.CorpID != "corp-test-001" {
		t.Fatalf("corp id mismatch: %q", cfg.CorpID)
	}
	if cfg.Secret != "11111111111111111111111111111111" {
		t.Fatalf("secret mismatch: %q", cfg.Secret)
	}
	if cfg.Account != "tester@example.com" {
		t.Fatalf("account mismatch: %q", cfg.Account)
	}
	if cfg.ApplicationID != "aaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("application id mismatch: %q", cfg.ApplicationID)
	}
	if cfg.FormModelID != "bbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("form model id mismatch: %q", cfg.FormModelID)
	}
}

func TestBuildSaveRecordPayloadKeepsVariablesSeparate(t *testing.T) {
	payload := BuildSaveRecordPayload(map[string]any{
		"测试文本":        "hello",
		"id":          "existing-id",
		"loginUserId": "wrong-user",
		"version":     7,
	}, "doc-1", "user-1")

	want := map[string]any{
		"id":          "doc-1",
		"loginUserId": "user-1",
		"variables": map[string]any{
			"测试文本": "hello",
		},
	}
	if !reflect.DeepEqual(payload, want) {
		t.Fatalf("payload mismatch:\nwant=%#v\n got=%#v", want, payload)
	}
}

func TestBuildUpdateRecordPayloadKeepsCurrentVersion(t *testing.T) {
	payload := BuildUpdateRecordPayload(map[string]any{"状态": "已更新"}, "doc-1", "user-1", 2)

	if payload["id"] != "doc-1" || payload["loginUserId"] != "user-1" || payload["version"] != 2 {
		t.Fatalf("unexpected update metadata: %#v", payload)
	}
	variables, ok := payload["variables"].(map[string]any)
	if !ok || variables["状态"] != "已更新" {
		t.Fatalf("unexpected variables: %#v", payload["variables"])
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	got := NormalizeBaseURL("http://intranet-qiqiao.example.local/qiqiao/runtime/api/v1/bpms-integration///")
	want := "http://intranet-qiqiao.example.local/qiqiao/runtime/api/v1/bpms-integration"
	if got != want {
		t.Fatalf("NormalizeBaseURL() = %q, want %q", got, want)
	}
}

func TestLoadConfigDoesNotInjectDeploymentBaseURL(t *testing.T) {
	cfg, err := LoadConfig("", "")
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.BaseURL != "" {
		t.Fatalf("base URL should be explicitly configured, got %q", cfg.BaseURL)
	}
	if err := validateAuthConfig(Config{CorpID: "corp", Secret: "11111111111111111111111111111111", Account: "tester"}); err == nil {
		t.Fatalf("validateAuthConfig should require an explicit base URL")
	}
}

func TestDecodeJSONBodyAcceptsCompleteJSONWithUnexpectedEOF(t *testing.T) {
	var env APIEnvelope
	err := decodeJSONBody([]byte(`{"code":0,"msg":"执行成功","data":[]}`), io.ErrUnexpectedEOF, &env, "test")
	if err != nil {
		t.Fatalf("decodeJSONBody returned error: %v", err)
	}
	if !envelopeOK(env) || env.Msg != "执行成功" {
		t.Fatalf("unexpected envelope: %#v", env)
	}
}

func TestVersionFromEnvelopeReadsCreateResponseVersion(t *testing.T) {
	env := APIEnvelope{Data: []byte(`{"id":"doc-1","version":3}`)}
	version, ok := versionFromEnvelope(env)
	if !ok || version != 3 {
		t.Fatalf("versionFromEnvelope() = %d, %v; want 3, true", version, ok)
	}
}

func TestSmokeVariablesUseKnownFieldTypes(t *testing.T) {
	fields := []FormField{
		{Name: "单行文本1", Type: "textBox"},
		{Name: "多行文本1", Type: "textarea"},
		{Name: "时间1", Type: "time"},
		{Name: "日期1", Type: "date"},
		{Name: "文件上传1", Type: "fileupload"},
	}
	vars := BuildSmokeVariables(fields, "smoke-001")

	if vars["单行文本1"] == "" || vars["多行文本1"] == "" {
		t.Fatalf("text fields were not populated: %#v", vars)
	}
	if vars["时间1"] == "" {
		t.Fatalf("time field was not populated: %#v", vars)
	}
	if _, ok := vars["日期1"].(int64); !ok {
		t.Fatalf("date field should be millisecond timestamp: %#v", vars["日期1"])
	}
	if _, exists := vars["文件上传1"]; exists {
		t.Fatalf("file upload field should not be populated by generic variables: %#v", vars)
	}
}

func TestStepFromEnvelopeMarksBusinessFailure(t *testing.T) {
	step := stepFromEnvelope("form_models", APIEnvelope{Code: float64(-1), Msg: "系统繁忙，请稍后再试"}, nil)
	if step.OK {
		t.Fatalf("business failure should not be ok: %#v", step)
	}
	if step.Message != "系统繁忙，请稍后再试" {
		t.Fatalf("unexpected message: %#v", step)
	}
}

func TestFormDesignProbeNoteIsExplicitlyNonDestructive(t *testing.T) {
	note := formDesignProbeNote()
	if !strings.Contains(note, "non-destructive") || !strings.Contains(note, "not create form models") {
		t.Fatalf("probe note should describe non-destructive boundary: %q", note)
	}
}

func TestBuildBatchUpdatePayloadIncrementsVersionForBatchEndpoint(t *testing.T) {
	create := APIEnvelope{Data: []byte(`[{"id":"doc-1","version":1}]`)}
	payload := buildBatchUpdatePayload(create, []FormField{{Name: "单行文本1", Type: "textBox"}}, "user-1")
	if len(payload) != 1 {
		t.Fatalf("expected one update payload, got %#v", payload)
	}
	if payload[0]["version"] != 2 {
		t.Fatalf("batch update should use version+1, got %#v", payload[0])
	}
}

func TestFullSmokeOKIgnoresOptionalDiscoveryAndDesignBoundary(t *testing.T) {
	steps := []TestStep{
		{Name: "applications.list", OK: false},
		{Name: "form_models.list", OK: false},
		{Name: "files.upload", OK: false},
		{Name: "form_design.create_form_or_fields", OK: false},
		{Name: "forms.create_one", OK: true},
	}
	if !fullSmokeOK(steps) {
		t.Fatalf("optional discovery/design boundary should not fail full smoke")
	}
}
