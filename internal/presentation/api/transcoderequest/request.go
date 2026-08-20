package transcoderequest

type CreateTranscodeRequestReq struct {
	VideoURL string `json:"video_url" validate:"required,url"`
}
