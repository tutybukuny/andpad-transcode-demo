package constant

type TranscodeRequestStatus string

const (
	TranscodeRequestStatusTodo       TranscodeRequestStatus = "todo"
	TranscodeRequestStatusProcessing TranscodeRequestStatus = "processing"
	TranscodeRequestStatusCompleted  TranscodeRequestStatus = "completed"
	TranscodeRequestStatusFailed     TranscodeRequestStatus = "failed"
	TranscodeRequestStatusCancelled  TranscodeRequestStatus = "cancelled"
)
