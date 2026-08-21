package transcoderequest

type CreateTranscodeRequestReq struct {
	VideoURL string `json:"video_url" validate:"required,url"`
}

type GetTranscodeRequestReq struct {
	ID int64 `uri:"id" validate:"required,gt=0"`
}

type CancelTranscodeRequestReq struct {
	ID int64 `uri:"id" validate:"required,gt=0"`
}
