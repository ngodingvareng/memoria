//go:build unit

package usecase_test

// strPtr is a small literal-to-pointer helper shared by tests across this
// package for optional entity.User fields like Username.
func strPtr(s string) *string { return &s }
