package notarize

// rawError is the error struct type for plist values.
type rawError struct {
	Code     int64             `plist:"code"`
	Message  string            `plist:"message"`
	UserInfo map[string]string `plist:"userInfo"`
}
