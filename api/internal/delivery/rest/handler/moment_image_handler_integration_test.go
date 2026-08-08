//go:build integration

package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

func createTestMoment(t *testing.T, testApp *testApp, userID uuid.UUID) dto.MomentResponse {
	t.Helper()
	resp := testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return decodeBody[dto.WebResponse[dto.MomentResponse]](t, resp).Data
}

func TestMomentImageHandler_UploadMomentImage_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, userID)

	resp := testApp.doMultipartUpload(t, "/moments/"+moment.ID+"/images", "image", "photo.png", validPNGBytes(t),
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.MomentImageResponse]](t, resp)
	require.NotEmpty(t, body.Data.ID)
	require.True(t, strings.HasPrefix(body.Data.URL, "https://fake.storage.test/"))
}

func TestMomentImageHandler_UploadMomentImage_NoFile(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, userID)

	resp := testApp.doMultipartUploadNoFile(t, "/moments/"+moment.ID+"/images",
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMomentImageHandler_UploadMomentImage_NotAnImage(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, userID)

	resp := testApp.doMultipartUpload(t, "/moments/"+moment.ID+"/images", "image", "note.txt", []byte("not an image"),
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMomentImageHandler_UploadMomentImage_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, userID)

	resp := testApp.doMultipartUpload(t, "/moments/"+moment.ID+"/images", "image", "photo.png", validPNGBytes(t), nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMomentImageHandler_UploadMomentImage_WrongOwner_NotFound(t *testing.T) {
	// MomentAccessChecker.GetByID returns errs.ErrNotFound for a
	// non-owner (moment_repository.go's GetByID doc comment).
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doMultipartUpload(t, "/moments/"+moment.ID+"/images", "image", "photo.png", validPNGBytes(t),
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMomentImageHandler_ListMomentImages_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, userID)

	for range 2 {
		uploadResp := testApp.doMultipartUpload(t, "/moments/"+moment.ID+"/images", "image", "photo.png", validPNGBytes(t),
			map[string]string{"Authorization": testApp.authHeader(t, userID)})
		require.Equal(t, http.StatusCreated, uploadResp.StatusCode)
	}

	resp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/images", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[[]dto.MomentImageResponse]](t, resp)
	require.Len(t, body.Data, 2)
}

func TestMomentImageHandler_DeleteMomentImage_ThenList_Gone(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, userID)

	uploaded := decodeBody[dto.WebResponse[dto.MomentImageResponse]](t,
		testApp.doMultipartUpload(t, "/moments/"+moment.ID+"/images", "image", "photo.png", validPNGBytes(t),
			map[string]string{"Authorization": testApp.authHeader(t, userID)}))

	deleteResp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/images/"+uploaded.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	listResp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/images", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	listBody := decodeBody[dto.WebResponse[[]dto.MomentImageResponse]](t, listResp)
	require.Empty(t, listBody.Data)
}

func TestMomentImageHandler_DeleteMomentImage_ForeignImageID_NoOp(t *testing.T) {
	// MomentImageRepository.Delete's underlying query is :exec — a WHERE
	// clause matching zero rows (unknown imageId) is a silent no-op, same
	// as ThreadImageRepository.Delete.
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	moment := createTestMoment(t, testApp, userID)

	resp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/images/"+uuid.NewString(), nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
