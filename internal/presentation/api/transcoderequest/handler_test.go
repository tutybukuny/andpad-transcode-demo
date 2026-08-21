package transcoderequest_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"transcode-demo/config"
	"transcode-demo/internal/constant"
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
			require.Equal(t, te.expectedStatusCode, resp.StatusCode())
			if te.assertFunc != nil {
				te.assertFunc(t, te.videoURL, &result)
			}
		})
	}
}
