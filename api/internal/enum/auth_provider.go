package enum

type AuthProvider string

const (
	AuthProviderGoogle     AuthProvider = "google"
	AuthProviderGitHub     AuthProvider = "github"
	AuthProviderCredential AuthProvider = "credential"
)
