package elasticsearch_api_users

type UserAPISpec struct {
	PasswordHash string `json:"hash"`
}
