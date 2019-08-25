package memobird

// APIResult contains the result of every API request.
type APIResult struct {
	IsSuccess bool
	Err       error
}
