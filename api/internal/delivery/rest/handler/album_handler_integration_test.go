//go:build integration

package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// createTestMomentAt is createTestMoment with a caller-chosen occurred_at,
// for tests that assert ordering.
func createTestMomentAt(t *testing.T, testApp *testApp, auth, occurredAt string) dto.MomentResponse {
	t.Helper()
	resp := testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: occurredAt,
	}, map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return decodeBody[dto.WebResponse[dto.MomentResponse]](t, resp).Data
}

func uploadTestImage(t *testing.T, testApp *testApp, momentID, auth string) dto.MomentImageResponse {
	t.Helper()
	resp := testApp.doMultipartUpload(t, "/moments/"+momentID+"/images", "image", "photo.png", validPNGBytes(t),
		map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return decodeBody[dto.WebResponse[dto.MomentImageResponse]](t, resp).Data
}

func TestAlbumHandler_ListAlbum_ReturnsOwnImagesOnly(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	otherID := seedTestUser(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)
	otherAuth := testApp.authHeader(t, otherID)

	momentWithImage := createTestMoment(t, testApp, ownerID)
	uploadTestImage(t, testApp, momentWithImage.ID, ownerAuth)
	createTestMoment(t, testApp, ownerID) // moment with no image — must not appear

	otherMoment := createTestMoment(t, testApp, otherID)
	uploadTestImage(t, testApp, otherMoment.ID, otherAuth)

	resp := testApp.doRequest(t, http.MethodGet, "/album", nil, map[string]string{"Authorization": ownerAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, resp)
	require.Len(t, body.Data.Images, 1)
	require.Equal(t, momentWithImage.ID, body.Data.Images[0].MomentID)
}

func TestAlbumHandler_ListAlbum_NewestOccurrenceFirst(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, userID)

	older := createTestMomentAt(t, testApp, auth, "2020-01-01T08:00:00Z")
	newer := createTestMomentAt(t, testApp, auth, "2026-01-01T08:00:00Z")
	uploadTestImage(t, testApp, older.ID, auth)
	uploadTestImage(t, testApp, newer.ID, auth)

	resp := testApp.doRequest(t, http.MethodGet, "/album", nil, map[string]string{"Authorization": auth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, resp)
	require.Len(t, body.Data.Images, 2)
	require.Equal(t, newer.ID, body.Data.Images[0].MomentID)
	require.Equal(t, older.ID, body.Data.Images[1].MomentID)
}

func TestAlbumHandler_ListAlbum_Pagination(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, userID)

	for range 3 {
		moment := createTestMoment(t, testApp, userID)
		uploadTestImage(t, testApp, moment.ID, auth)
	}

	page1 := testApp.doRequest(t, http.MethodGet, "/album?page=1&page_size=2", nil, map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusOK, page1.StatusCode)
	page1Body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, page1)
	require.Len(t, page1Body.Data.Images, 2)

	page2 := testApp.doRequest(t, http.MethodGet, "/album?page=2&page_size=2", nil, map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusOK, page2.StatusCode)
	page2Body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, page2)
	require.Len(t, page2Body.Data.Images, 1)
}

func TestAlbumHandler_ListAlbum_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodGet, "/album", nil, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAlbumHandler_ListCircleAlbum_UnionsNativeAndSharedImages(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)

	circle := createCircle(t, testApp, ownerAuth, "Family")

	// Native: a Moment captured directly into the Circle's own Thread.
	thread := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
			Name:     "Family trips",
			CircleID: &circle.ID,
		}, map[string]string{"Authorization": ownerAuth})).Data
	nativeMoment := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			ThreadID:   &thread.ID,
			OccurredAt: "2026-08-04T14:30:00+09:00",
		}, map[string]string{"Authorization": ownerAuth})).Data
	uploadTestImage(t, testApp, nativeMoment.ID, ownerAuth)

	// Shared: a personal Moment shared into the Circle via the mention flow.
	sharedMoment := createTestMoment(t, testApp, ownerID)
	uploadTestImage(t, testApp, sharedMoment.ID, ownerAuth)
	shareResp := testApp.doRequest(t, http.MethodPost, "/moments/"+sharedMoment.ID+"/share",
		dto.ShareMomentToCircleRequest{CircleID: circle.ID},
		map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, shareResp.StatusCode)

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/album", nil,
		map[string]string{"Authorization": ownerAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, resp)
	require.Len(t, body.Data.Images, 2)
	momentIDs := []string{body.Data.Images[0].MomentID, body.Data.Images[1].MomentID}
	require.ElementsMatch(t, []string{nativeMoment.ID, sharedMoment.ID}, momentIDs)
}

func TestAlbumHandler_ListCircleAlbum_SharedImageCarriesAttribution_NativeDoesNot(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID, ownerUsername := seedTestUserWithUsername(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)

	circle := createCircle(t, testApp, ownerAuth, "Family")

	thread := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
			Name:     "Family trips",
			CircleID: &circle.ID,
		}, map[string]string{"Authorization": ownerAuth})).Data
	nativeMoment := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			ThreadID:   &thread.ID,
			OccurredAt: "2026-08-04T14:30:00+09:00",
		}, map[string]string{"Authorization": ownerAuth})).Data
	uploadTestImage(t, testApp, nativeMoment.ID, ownerAuth)

	sharedMoment := createTestMoment(t, testApp, ownerID)
	uploadTestImage(t, testApp, sharedMoment.ID, ownerAuth)
	shareResp := testApp.doRequest(t, http.MethodPost, "/moments/"+sharedMoment.ID+"/share",
		dto.ShareMomentToCircleRequest{CircleID: circle.ID},
		map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, shareResp.StatusCode)

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/album", nil,
		map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, resp)
	require.Len(t, body.Data.Images, 2)

	byMoment := map[string]dto.AlbumImageResponse{}
	for _, img := range body.Data.Images {
		byMoment[img.MomentID] = img
	}

	native := byMoment[nativeMoment.ID]
	require.False(t, native.IsShared)
	require.Nil(t, native.SharedBy)

	shared := byMoment[sharedMoment.ID]
	require.True(t, shared.IsShared)
	require.NotNil(t, shared.SharedBy)
	require.Equal(t, ownerID.String(), shared.SharedBy.ID)
	require.NotNil(t, shared.SharedBy.Username)
	require.Equal(t, ownerUsername, *shared.SharedBy.Username)
}

func TestAlbumHandler_ListCircleAlbum_NativeMomentAlsoSharedIntoSameCircle_NotDuplicated_NativeWins(t *testing.T) {
	// Regression test: ShareToCircle only checks moment ownership + circle
	// membership, not "unless already native to this circle" — so a Moment
	// captured directly into the Circle's own Thread can also pick up a
	// moment_circles row for that same Circle. The Album query must not
	// show the image twice, and the surviving row must be unattributed
	// (native wins over shared).
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)

	circle := createCircle(t, testApp, ownerAuth, "Family")

	thread := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
			Name:     "Family trips",
			CircleID: &circle.ID,
		}, map[string]string{"Authorization": ownerAuth})).Data
	nativeMoment := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			ThreadID:   &thread.ID,
			OccurredAt: "2026-08-04T14:30:00+09:00",
		}, map[string]string{"Authorization": ownerAuth})).Data
	uploadTestImage(t, testApp, nativeMoment.ID, ownerAuth)

	shareResp := testApp.doRequest(t, http.MethodPost, "/moments/"+nativeMoment.ID+"/share",
		dto.ShareMomentToCircleRequest{CircleID: circle.ID},
		map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, shareResp.StatusCode)

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/album", nil,
		map[string]string{"Authorization": ownerAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, resp)
	require.Len(t, body.Data.Images, 1)
	require.False(t, body.Data.Images[0].IsShared)
	require.Nil(t, body.Data.Images[0].SharedBy)
}

func TestAlbumHandler_ListCircleAlbum_NotMember_NotFound(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	circle := createCircle(t, testApp, testApp.authHeader(t, ownerID), "Family")

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/album", nil,
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAlbumHandler_ListCircleAlbum_Pagination(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)

	circle := createCircle(t, testApp, ownerAuth, "Family")
	thread := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
			Name:     "Family trips",
			CircleID: &circle.ID,
		}, map[string]string{"Authorization": ownerAuth})).Data

	for range 3 {
		moment := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
			testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
				ThreadID:   &thread.ID,
				OccurredAt: "2026-08-04T14:30:00+09:00",
			}, map[string]string{"Authorization": ownerAuth})).Data
		uploadTestImage(t, testApp, moment.ID, ownerAuth)
	}

	page1 := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/album?page=1&page_size=2", nil,
		map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, page1.StatusCode)
	page1Body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, page1)
	require.Len(t, page1Body.Data.Images, 2)

	page2 := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/album?page=2&page_size=2", nil,
		map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, page2.StatusCode)
	page2Body := decodeBody[dto.WebResponse[dto.ListAlbumResponse]](t, page2)
	require.Len(t, page2Body.Data.Images, 1)
}
