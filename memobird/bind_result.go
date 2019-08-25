package memobird

// BindResult contains the result of a bind request.
type BindResult struct {
	APIResult
	UserID uint64
}
