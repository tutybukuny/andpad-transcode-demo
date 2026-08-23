package transcoderequest_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
	"transcode-demo/internal/models/entity"
	"transcode-demo/internal/models/response"
	"transcode-demo/internal/presentation/api"
	"transcode-demo/internal/presentation/api/transcoderequest"
	"transcode-demo/internal/repositories/psqlrepo"
	"transcode-demo/pkg/cerrors"
	"transcode-demo/pkg/utils"
)

func TestHandler_CreateTranscodeRequest(t *testing.T) {
	ctx := context.Background()
	db := config.MustNewGorm(t)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	serverURL := api.NewTestServer(t)
	client := resty.New().SetBaseURL(serverURL)

	assertBadRequest := func(t *testing.T, _ string, result *response.Response[transcoderequest.CreateTranscodeRequestResp]) {
		require.Empty(t, result.Data)
		require.Equal(t, cerrors.BadRequest, result.Code)
		require.Equal(t, cerrors.BadRequest.String(), result.Status)
	}
	assertInvalidField := func(t *testing.T, inputtedVideoURL string, result *response.Response[transcoderequest.CreateTranscodeRequestResp]) {
		assertBadRequest(t, inputtedVideoURL, result)
		require.Contains(t, result.Message, "Field validation for 'VideoURL' failed")
	}

	for _, te := range []struct {
		name               string
		videoURL           string
		customBody         string
		expectedStatusCode int
		assertFunc         func(t *testing.T, inputtedVideoURL string, result *response.Response[transcoderequest.CreateTranscodeRequestResp])
	}{
		{
			name:               "create_successfully",
			videoURL:           fmt.Sprintf("s3://local-storage/%s.mp4", utils.RandString()),
			expectedStatusCode: 200,
			assertFunc: func(t *testing.T, inputtedVideoURL string, result *response.Response[transcoderequest.CreateTranscodeRequestResp]) {
				require.Equal(t, cerrors.OK, result.Code)
				require.Equal(t, cerrors.OK.String(), result.Status)
				require.NotEmpty(t, result.Data.RequestID)
				req, err := transReqRepo.FindByID(ctx, result.Data.RequestID)
				require.NoError(t, err)
				require.Equal(t, inputtedVideoURL, req.VideoURL)
				require.Equal(t, constant.TranscodeRequestStatusTodo, req.Status)
			},
		},
		{
			name:               "invalid_video_url",
			videoURL:           utils.RandString(),
			expectedStatusCode: 400,
			assertFunc:         assertInvalidField,
		},
		{
			name:               "empty_video_url",
			expectedStatusCode: 400,
			assertFunc:         assertInvalidField,
		},
		{
			name:               "invalid_body",
			expectedStatusCode: 400,
			customBody:         "not a valid json",
			assertFunc: func(t *testing.T, inputtedVideoURL string, result *response.Response[transcoderequest.CreateTranscodeRequestResp]) {
				assertBadRequest(t, inputtedVideoURL, result)
				require.Contains(t, result.Message, "Unprocessable Entity")
			},
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			var result response.Response[transcoderequest.CreateTranscodeRequestResp]
			var body any = transcoderequest.CreateTranscodeRequestReq{VideoURL: te.videoURL}
			if te.customBody != "" {
				body = te.customBody
			}
			resp, err := client.R().
				SetBody(body).
				SetResult(&result).
				SetError(&result).
				Post("/transcode-request")
			require.NoError(t, err)
			if te.expectedStatusCode == 200 {
				t.Cleanup(func() {
					err = transReqRepo.Delete(ctx, result.Data.RequestID)
					require.NoError(t, err)
				})
			}
			require.Equal(t, te.expectedStatusCode, resp.StatusCode())
			if te.assertFunc != nil {
				te.assertFunc(t, te.videoURL, &result)
			}
		})
	}
}

func TestHandler_GetTranscodeRequest(t *testing.T) {
	ctx := context.Background()
	db := config.MustNewGorm(t)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	serverURL := api.NewTestServer(t)
	client := resty.New().SetBaseURL(serverURL)

	// fixture a request
	transReq := &entity.TranscodeRequest{
		VideoURL:            utils.RandString(),
		OutputURL:           utils.RandString(),
		Status:              constant.TranscodeRequestStatusCompleted,
		StartedTranscodeAt:  new(time.Now().Add(-10 * time.Minute)),
		FinishedTranscodeAt: new(time.Now().Add(-5 * time.Minute)),
	}
	err := transReqRepo.Insert(ctx, transReq)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = transReqRepo.Delete(ctx, transReq.ID)
		require.NoError(t, err)
	})

	assertBadRequest := func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
		require.Empty(t, result.Data)
		require.Equal(t, cerrors.BadRequest, result.Code)
		require.Equal(t, cerrors.BadRequest.String(), result.Status)
	}

	for _, te := range []struct {
		name               string
		transID            string
		expectedStatusCode int
		assertFunc         func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp])
	}{
		{
			name:               "get_successfully",
			transID:            fmt.Sprintf("%d", transReq.ID),
			expectedStatusCode: 200,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				require.Equal(t, cerrors.OK, result.Code)
				require.Equal(t, cerrors.OK.String(), result.Status)
				require.Equal(t, transReq.ID, result.Data.TranscodeRequest.ID)
				require.Equal(t, transReq.VideoURL, result.Data.TranscodeRequest.VideoURL)
				require.Equal(t, transReq.OutputURL, result.Data.TranscodeRequest.OutputURL)
				require.Equal(t, transReq.Status, result.Data.TranscodeRequest.Status)
				require.Equal(t, transReq.StartedTranscodeAt.Unix(), result.Data.TranscodeRequest.StartedTranscodeAt.Unix())
				require.Equal(t, transReq.FinishedTranscodeAt.Unix(), result.Data.TranscodeRequest.FinishedTranscodeAt.Unix())
			},
		},
		{
			name:               "invalid_trans_id",
			transID:            utils.RandString(),
			expectedStatusCode: 400,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				assertBadRequest(t, result)
				require.Contains(t, result.Message, `converting value for "id"`)
			},
		},
		{
			name:               "negative_trans_id",
			transID:            "-123",
			expectedStatusCode: 400,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				assertBadRequest(t, result)
				require.Contains(t, result.Message, "Field validation for 'ID'")
			},
		},
		{
			name:               "not_found_trans_id",
			transID:            "1234567890",
			expectedStatusCode: 404,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				require.Empty(t, result.Data)
				require.Equal(t, cerrors.Code(404001), result.Code)
				require.Equal(t, cerrors.NotFound.String(), result.Status)
				require.Equal(t, "transcode request not found", result.Message)
			},
		},
		{
			name:               "empty_trans_id",
			expectedStatusCode: 405,
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			var result response.Response[transcoderequest.GetTranscodeRequestResp]
			var resp *resty.Response
			resp, err = client.R().
				SetResult(&result).
				SetError(&result).
				Get(fmt.Sprintf("/transcode-request/%s", te.transID))
			require.NoError(t, err)
			require.Equal(t, te.expectedStatusCode, resp.StatusCode())
			if te.assertFunc != nil {
				te.assertFunc(t, &result)
			}
		})
	}
}

func TestHandler_CancelTranscodeRequest(t *testing.T) {
	ctx := context.Background()
	db := config.MustNewGorm(t)
	transReqRepo := psqlrepo.NewTranscodeRequestRepo(db)
	serverURL := api.NewTestServer(t)
	client := resty.New().SetBaseURL(serverURL)

	assertBadRequest := func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
		require.Empty(t, result.Data)
		require.Equal(t, cerrors.BadRequest, result.Code)
		require.Equal(t, cerrors.BadRequest.String(), result.Status)
	}
	assertOKRequest := func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
		require.Equal(t, cerrors.OK, result.Code)
		require.Equal(t, cerrors.OK.String(), result.Status)
	}

	for _, te := range []struct {
		name                 string
		transID              *string
		transReqStatus       constant.TranscodeRequestStatus
		expectedStatusCode   int
		expectedEndingStatus constant.TranscodeRequestStatus
		assertFunc           func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp])
	}{
		{
			name:                 "cancel_todo",
			transReqStatus:       constant.TranscodeRequestStatusTodo,
			expectedStatusCode:   200,
			expectedEndingStatus: constant.TranscodeRequestStatusCancelled,
			assertFunc:           assertOKRequest,
		},
		{
			name:                 "cancel_processing",
			transReqStatus:       constant.TranscodeRequestStatusProcessing,
			expectedStatusCode:   200,
			expectedEndingStatus: constant.TranscodeRequestStatusCancelled,
			assertFunc:           assertOKRequest,
		},
		{
			name:                 "cancel_already_cancelled",
			transReqStatus:       constant.TranscodeRequestStatusCancelled,
			expectedStatusCode:   200,
			expectedEndingStatus: constant.TranscodeRequestStatusCancelled,
			assertFunc:           assertOKRequest,
		},
		{
			name:                 "cancel_completed",
			transReqStatus:       constant.TranscodeRequestStatusCompleted,
			expectedStatusCode:   400,
			expectedEndingStatus: constant.TranscodeRequestStatusCompleted,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				assertBadRequest(t, result)
				require.Contains(t, result.Message, "transcode request already completed")
			},
		},
		{
			name:                 "cancel_failed",
			transReqStatus:       constant.TranscodeRequestStatusFailed,
			expectedStatusCode:   400,
			expectedEndingStatus: constant.TranscodeRequestStatusFailed,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				assertBadRequest(t, result)
				require.Contains(t, result.Message, "transcode request already completed")
			},
		},
		{
			name:               "invalid_trans_id",
			transID:            new(utils.RandString()),
			expectedStatusCode: 400,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				assertBadRequest(t, result)
				require.Contains(t, result.Message, `converting value for "id"`)
			},
		},
		{
			name:               "negative_trans_id",
			transID:            new("-123"),
			expectedStatusCode: 400,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				assertBadRequest(t, result)
				require.Contains(t, result.Message, "Field validation for 'ID'")
			},
		},
		{
			name:               "not_found_trans_id",
			transID:            new("1234567890"),
			expectedStatusCode: 404,
			assertFunc: func(t *testing.T, result *response.Response[transcoderequest.GetTranscodeRequestResp]) {
				require.Empty(t, result.Data)
				require.Equal(t, cerrors.Code(404001), result.Code)
				require.Equal(t, cerrors.NotFound.String(), result.Status)
				require.Equal(t, "transcode request not found", result.Message)
			},
		},
		{
			name:               "empty_trans_id",
			transID:            new(""),
			expectedStatusCode: 404,
		},
	} {
		t.Run(te.name, func(t *testing.T) {
			var transID string
			var req entity.TranscodeRequest
			if te.transID != nil {
				transID = *te.transID
			} else {
				req = entity.TranscodeRequest{
					VideoURL: utils.RandString(),
					Status:   te.transReqStatus,
				}
				err := transReqRepo.Insert(ctx, &req)
				require.NoError(t, err)
				t.Cleanup(func() {
					err = transReqRepo.Delete(ctx, req.ID)
					require.NoError(t, err)
				})
				transID = fmt.Sprintf("%d", req.ID)
			}

			var result response.Response[transcoderequest.GetTranscodeRequestResp]
			resp, err := client.R().
				SetResult(&result).
				SetError(&result).
				Post(fmt.Sprintf("/transcode-request/%s/cancel", transID))
			require.NoError(t, err)
			require.Equal(t, te.expectedStatusCode, resp.StatusCode())
			if te.assertFunc != nil {
				te.assertFunc(t, &result)
			}
			if te.transID != nil {
				return
			}
			r, err := transReqRepo.FindByID(ctx, req.ID)
			require.NoError(t, err)
			require.Equal(t, te.expectedEndingStatus, r.Status)
		})
	}
}
