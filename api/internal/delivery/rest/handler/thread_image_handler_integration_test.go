//go:build integration

package handler_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/require"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// doMultipartUpload sends a single-file multipart/form-data POST through
// the full route tree, mirroring doRequest but for uploads — doRequest
// only JSON-marshals its body, which can't express a file part.
func (a *testApp) doMultipartUpload(t *testing.T, path, fieldName, fileName string, fileContent []byte, headers map[string]string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)
	_, err = part.Write(fileContent)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := a.app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second, FailOnTimeout: true})
	require.NoError(t, err)
	return resp
}

// doMultipartUploadNoFile posts a multipart body with no file part at
// all, to exercise the "image file is required" branch — doMultipartUpload
// can't express "no part" since it always writes one.
func (a *testApp) doMultipartUploadNoFile(t *testing.T, path string, headers map[string]string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	require.NoError(t, w.WriteField("note", "no file here"))
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := a.app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second, FailOnTimeout: true})
	require.NoError(t, err)
	return resp
}

// validPNGBytes builds a real, minimal PNG at test time rather than
// hardcoding raw magic bytes, so http.DetectContentType sniffs it as
// image/png the same way it would a real upload.
func validPNGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func createTestThread(t *testing.T, testApp *testApp, userID uuid.UUID) dto.ThreadResponse {
	t.Helper()
	resp := testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Morning workout"},
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return decodeBody[dto.WebResponse[dto.ThreadResponse]](t, resp).Data
}

func TestThreadImageHandler_UploadThreadImage_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, userID)

	resp := testApp.doMultipartUpload(t, "/threads/"+thread.ID+"/images", "image", "avatar.png", validPNGBytes(t),
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ThreadImageResponse]](t, resp)
	require.NotEmpty(t, body.Data.ID)
	require.True(t, strings.HasPrefix(body.Data.URL, "https://fake.storage.test/"))
}

func TestThreadImageHandler_UploadThreadImage_NoFile(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, userID)

	resp := testApp.doMultipartUploadNoFile(t, "/threads/"+thread.ID+"/images",
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestThreadImageHandler_UploadThreadImage_NotAnImage(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, userID)

	resp := testApp.doMultipartUpload(t, "/threads/"+thread.ID+"/images", "image", "note.txt", []byte("not an image"),
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestThreadImageHandler_UploadThreadImage_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, userID)

	resp := testApp.doMultipartUpload(t, "/threads/"+thread.ID+"/images", "image", "avatar.png", validPNGBytes(t), nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestThreadImageHandler_UploadThreadImage_WrongOwner_NotFound(t *testing.T) {
	// ThreadAccessChecker.GetByID returns errs.ErrNotFound for a non-owner
	// (same "wrong owner indistinguishable from wrong id" rule as threads
	// themselves — see thread_usecase.go).
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, ownerID)

	resp := testApp.doMultipartUpload(t, "/threads/"+thread.ID+"/images", "image", "avatar.png", validPNGBytes(t),
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestThreadImageHandler_ListThreadImages_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, userID)

	for range 2 {
		uploadResp := testApp.doMultipartUpload(t, "/threads/"+thread.ID+"/images", "image", "avatar.png", validPNGBytes(t),
			map[string]string{"Authorization": testApp.authHeader(t, userID)})
		require.Equal(t, http.StatusCreated, uploadResp.StatusCode)
	}

	resp := testApp.doRequest(t, http.MethodGet, "/threads/"+thread.ID+"/images", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[[]dto.ThreadImageResponse]](t, resp)
	require.Len(t, body.Data, 2)
}

func TestThreadImageHandler_DeleteThreadImage_ThenList_Gone(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, userID)

	uploaded := decodeBody[dto.WebResponse[dto.ThreadImageResponse]](t,
		testApp.doMultipartUpload(t, "/threads/"+thread.ID+"/images", "image", "avatar.png", validPNGBytes(t),
			map[string]string{"Authorization": testApp.authHeader(t, userID)}))

	deleteResp := testApp.doRequest(t, http.MethodDelete, "/threads/"+thread.ID+"/images/"+uploaded.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	listResp := testApp.doRequest(t, http.MethodGet, "/threads/"+thread.ID+"/images", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	listBody := decodeBody[dto.WebResponse[[]dto.ThreadImageResponse]](t, listResp)
	require.Empty(t, listBody.Data)
}

func TestThreadImageHandler_DeleteThreadImage_ForeignImageID_NoOp(t *testing.T) {
	// ThreadImageRepository.Delete's underlying query is :exec — a WHERE
	// clause matching zero rows (unknown imageId) is a silent no-op, not
	// errs.ErrNotFound (mirrors ThreadRepository.SoftDelete, see
	// thread_usecase.go).
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	thread := createTestThread(t, testApp, userID)

	resp := testApp.doRequest(t, http.MethodDelete, "/threads/"+thread.ID+"/images/"+uuid.NewString(), nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
