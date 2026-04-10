package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/config"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/database"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/router"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	seedProjectID      = "33333333-3333-3333-3333-333333333333"
	seedSecondaryUser  = "22222222-2222-2222-2222-222222222222"
	seedPrimaryEmail   = "test@example.com"
	seedSecondaryEmail = "collab@example.com"
	seedPassword       = "password123"
)

var require requirePkg

type testHarness struct {
	baseURL string
	client  *http.Client
}

func TestAuthHappyPath_RegisterLoginAndProtectedAccess(t *testing.T) {
	t.Parallel()

	h := setupHarness(t)
	email := fmt.Sprintf("integration-auth-%d@example.com", time.Now().UnixNano())

	registerPayload := map[string]any{
		"name":     "Integration User",
		"email":    email,
		"password": "password123",
	}
	registerStatus, registerBody := h.jsonRequest(t, http.MethodPost, "/auth/register", "", registerPayload)
	require.Equal(t, http.StatusCreated, registerStatus)
	require.NotEmpty(t, asString(registerBody["token"]))

	loginPayload := map[string]any{
		"email":    email,
		"password": "password123",
	}
	loginStatus, loginBody := h.jsonRequest(t, http.MethodPost, "/auth/login", "", loginPayload)
	require.Equal(t, http.StatusOK, loginStatus)
	token := asString(loginBody["token"])
	require.NotEmpty(t, token)

	meStatus, meBody := h.jsonRequest(t, http.MethodGet, "/me", token, nil)
	require.Equal(t, http.StatusOK, meStatus)
	user, ok := meBody["user"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, strings.ToLower(email), asString(user["email"]))
}

func TestAuthDistinction_MissingTokenIs401AndForbiddenMutationIs403(t *testing.T) {
	t.Parallel()

	h := setupHarness(t)

	missingTokenStatus, missingTokenBody := h.jsonRequest(t, http.MethodGet, "/projects", "", nil)
	require.Equal(t, http.StatusUnauthorized, missingTokenStatus)
	require.Equal(t, "unauthorized", asString(missingTokenBody["error"]))

	collabToken := h.loginWithSeedUser(t, seedSecondaryEmail, seedPassword)
	patchPayload := map[string]any{
		"name": "Should Not Work",
	}

	forbiddenStatus, forbiddenBody := h.jsonRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/projects/%s", seedProjectID),
		collabToken,
		patchPayload,
	)
	require.Equal(t, http.StatusForbidden, forbiddenStatus)
	require.Equal(t, "forbidden", asString(forbiddenBody["error"]))
}

func TestTaskFlow_CreateUpdateFilterPaginateStatsAndDeleteRules(t *testing.T) {
	t.Parallel()

	h := setupHarness(t)

	ownerToken := h.loginWithSeedUser(t, seedPrimaryEmail, seedPassword)

	createPayload := map[string]any{
		"title":       "Integration Task",
		"description": "Created by integration test",
		"status":      "todo",
		"priority":    "high",
		"assignee_id": seedSecondaryUser,
		"due_date":    "2026-12-31",
	}
	createStatus, createBody := h.jsonRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/projects/%s/tasks", seedProjectID),
		ownerToken,
		createPayload,
	)
	require.Equal(t, http.StatusCreated, createStatus)
	createdTaskID := asString(createBody["id"])
	require.NotEmpty(t, createdTaskID)
	require.Equal(t, "todo", asString(createBody["status"]))

	updatePayload := map[string]any{
		"status":   "done",
		"priority": "medium",
	}
	updateStatus, updateBody := h.jsonRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/tasks/%s", createdTaskID),
		ownerToken,
		updatePayload,
	)
	require.Equal(t, http.StatusOK, updateStatus)
	require.Equal(t, "done", asString(updateBody["status"]))
	require.Equal(t, "medium", asString(updateBody["priority"]))

	filterStatus, filterBody := h.jsonRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/projects/%s/tasks?status=done&assignee=%s&page=1&limit=10", seedProjectID, seedSecondaryUser),
		ownerToken,
		nil,
	)
	require.Equal(t, http.StatusOK, filterStatus)
	filteredTasks := asObjectList(filterBody["tasks"])
	require.NotEmpty(t, filteredTasks)
	require.True(t, containsTask(filteredTasks, createdTaskID))

	pageOneStatus, pageOneBody := h.jsonRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/projects/%s/tasks?page=1&limit=2", seedProjectID),
		ownerToken,
		nil,
	)
	require.Equal(t, http.StatusOK, pageOneStatus)
	pageOneTasks := asObjectList(pageOneBody["tasks"])
	require.Len(t, pageOneTasks, 2)
	pageOneMeta := asObject(pageOneBody["pagination"])
	require.EqualValues(t, 1, asInt(pageOneMeta["page"]))
	require.EqualValues(t, 2, asInt(pageOneMeta["limit"]))
	require.GreaterOrEqual(t, asInt(pageOneMeta["total"]), 4)

	pageTwoStatus, pageTwoBody := h.jsonRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/projects/%s/tasks?page=2&limit=2", seedProjectID),
		ownerToken,
		nil,
	)
	require.Equal(t, http.StatusOK, pageTwoStatus)
	pageTwoTasks := asObjectList(pageTwoBody["tasks"])
	require.NotEmpty(t, pageTwoTasks)

	statsStatus, statsBody := h.jsonRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/projects/%s/stats", seedProjectID),
		ownerToken,
		nil,
	)
	require.Equal(t, http.StatusOK, statsStatus)
	byStatus := asObjectList(statsBody["by_status"])
	require.True(t, containsStatusCount(byStatus, "done"))

	outsiderEmail := fmt.Sprintf("integration-outsider-%d@example.com", time.Now().UnixNano())
	outsiderRegister := map[string]any{
		"name":     "Outsider",
		"email":    outsiderEmail,
		"password": "password123",
	}
	outsiderStatus, outsiderBody := h.jsonRequest(t, http.MethodPost, "/auth/register", "", outsiderRegister)
	require.Equal(t, http.StatusCreated, outsiderStatus)
	outsiderToken := asString(outsiderBody["token"])
	require.NotEmpty(t, outsiderToken)

	forbiddenOutsiderStatus, forbiddenOutsiderBody := h.jsonRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/tasks/%s", createdTaskID),
		outsiderToken,
		nil,
	)
	require.Equal(t, http.StatusForbidden, forbiddenOutsiderStatus)
	require.Equal(t, "forbidden", asString(forbiddenOutsiderBody["error"]))

	collabToken := h.loginWithSeedUser(t, seedSecondaryEmail, seedPassword)
	forbiddenCollabStatus, forbiddenCollabBody := h.jsonRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/tasks/%s", createdTaskID),
		collabToken,
		nil,
	)
	require.Equal(t, http.StatusForbidden, forbiddenCollabStatus)
	require.Equal(t, "forbidden", asString(forbiddenCollabBody["error"]))

	deleteStatus, _ := h.jsonRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/tasks/%s", createdTaskID),
		ownerToken,
		nil,
	)
	require.Equal(t, http.StatusNoContent, deleteStatus)

	deleteAgainStatus, deleteAgainBody := h.jsonRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/tasks/%s", createdTaskID),
		ownerToken,
		nil,
	)
	require.Equal(t, http.StatusNotFound, deleteAgainStatus)
	require.Equal(t, "not found", asString(deleteAgainBody["error"]))
}

func setupHarness(t *testing.T) *testHarness {
	t.Helper()

	testDatabaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if testDatabaseURL == "" {
		testDatabaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if testDatabaseURL == "" {
		t.Skip("integration test requires TEST_DATABASE_URL or DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, testDatabaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, resetDatabase(ctx, pool))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		ServerPort:  "0",
		DatabaseURL: testDatabaseURL,
		JWTSecret:   "integration-test-secret",
		JWTIssuer:   "taskflow-integration",
		JWTTTL:      24 * time.Hour,
		LogLevel:    slog.LevelError,
	}

	handler := router.New(logger, cfg, pool)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &testHarness{
		baseURL: server.URL,
		client:  server.Client(),
	}
}

func resetDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	if err := execSQL(ctx, pool, "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"); err != nil {
		return err
	}

	for _, file := range migrationFiles() {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		if err := execSQL(ctx, pool, string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", file, err)
		}
	}

	seedFile := filepath.Join(backendDir(), "seed", "001_seed.sql")
	seedContent, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read seed file: %w", err)
	}
	if err := execSQL(ctx, pool, string(seedContent)); err != nil {
		return fmt.Errorf("execute seed file: %w", err)
	}

	return nil
}

func execSQL(ctx context.Context, pool *pgxpool.Pool, script string) error {
	_, err := pool.Exec(ctx, script, pgx.QueryExecModeSimpleProtocol)
	return err
}

func migrationFiles() []string {
	files, err := filepath.Glob(filepath.Join(backendDir(), "migrations", "*.up.sql"))
	if err != nil {
		return nil
	}
	slices.Sort(files)
	return files
}

func backendDir() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), ".."))
}

func (h *testHarness) loginWithSeedUser(t *testing.T, email, password string) string {
	t.Helper()
	status, body := h.jsonRequest(t, http.MethodPost, "/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	require.Equal(t, http.StatusOK, status)
	token := asString(body["token"])
	require.NotEmpty(t, token)
	return token
}

func (h *testHarness) jsonRequest(t *testing.T, method, path, token string, payload any) (int, map[string]any) {
	t.Helper()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, h.baseURL+path, body)
	require.NoError(t, err)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := h.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return resp.StatusCode, map[string]any{}
	}

	rawBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	if len(bytes.TrimSpace(rawBody)) == 0 {
		return resp.StatusCode, map[string]any{}
	}

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(rawBody, &decoded))
	return resp.StatusCode, decoded
}

func asString(value any) string {
	s, _ := value.(string)
	return s
}

func asObject(value any) map[string]any {
	out, _ := value.(map[string]any)
	return out
}

func asObjectList(value any) []map[string]any {
	rawList, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(rawList))
	for _, item := range rawList {
		entry, ok := item.(map[string]any)
		if ok {
			out = append(out, entry)
		}
	}
	return out
}

func asInt(value any) int {
	number, ok := value.(float64)
	if !ok {
		return 0
	}
	return int(number)
}

func containsTask(tasks []map[string]any, taskID string) bool {
	for _, task := range tasks {
		if asString(task["id"]) == taskID {
			return true
		}
	}
	return false
}

func containsStatusCount(stats []map[string]any, status string) bool {
	for _, item := range stats {
		if asString(item["status"]) == status && asInt(item["count"]) > 0 {
			return true
		}
	}
	return false
}

type requirePkg struct{}

func (requirePkg) NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func (requirePkg) Equal(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func (requirePkg) EqualValues(t *testing.T, expected, actual any) {
	t.Helper()
	if fmt.Sprint(expected) != fmt.Sprint(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func (requirePkg) NotEmpty(t *testing.T, value any) {
	t.Helper()
	isEmpty := value == nil
	if !isEmpty {
		switch v := value.(type) {
		case string:
			isEmpty = strings.TrimSpace(v) == ""
		case []map[string]any:
			isEmpty = len(v) == 0
		default:
			isEmpty = reflect.ValueOf(value).Kind() == reflect.Slice && reflect.ValueOf(value).Len() == 0
		}
	}
	if isEmpty {
		t.Fatalf("expected non-empty value")
	}
}

func (requirePkg) True(t *testing.T, value bool) {
	t.Helper()
	if !value {
		t.Fatalf("expected condition to be true")
	}
}

func (requirePkg) Len(t *testing.T, list []map[string]any, expected int) {
	t.Helper()
	if len(list) != expected {
		t.Fatalf("expected length %d, got %d", expected, len(list))
	}
}

func (requirePkg) GreaterOrEqual(t *testing.T, actual int, expectedMin int) {
	t.Helper()
	if actual < expectedMin {
		t.Fatalf("expected %d to be >= %d", actual, expectedMin)
	}
}
