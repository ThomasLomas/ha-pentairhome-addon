package api

import "encoding/json"

type Profile struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ProfileResponse struct {
	Response Profile `json:"response"`
	Msgs     []any   `json:"msgs"`
}

func (client APIClient) GetProfile() Profile {
	body := client.MakeRequest("user/user-service/common/profile", "GET", nil)

	var result ProfileResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		panic("Can not unmarshal ProfileResponse JSON")
	}

	return result.Response
}
