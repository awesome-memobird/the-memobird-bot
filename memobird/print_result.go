package memobird

// PrintResult contains the result of a print request.
type PrintResult struct {
	APIResult
	IsPrinted bool
	ContentID int64
	DeviceID  string
}
