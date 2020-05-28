package fsm

// ApplyResponse response from Apply raft
type ApplyResponse struct {
	Error error
	Data  interface{}
}
