package str

import "encoding/base64"

func EncodedBase64(s string) string {
	data := []byte(s)
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeBase64(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
