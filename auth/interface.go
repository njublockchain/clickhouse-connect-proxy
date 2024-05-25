package auth

type AuthPlugin interface {
	Auth(apiToken string) bool
}
