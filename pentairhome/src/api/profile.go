package api

import (
	"encoding/json"
	"fmt"
)

type Profile struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ProfileResponse struct {
	Response Profile `json:"response"`
	Msgs     []any   `json:"msgs"`
}

func (client APIClient) GetProfile() (*Profile, error) {
	body, bodyErr := client.MakeRequest("user/user-service/common/profile", "GET", nil)

	if bodyErr != nil {
		return nil, fmt.Errorf("failed to get profile: %s", bodyErr)
	}

	var result ProfileResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		return nil, fmt.Errorf("failed to unmarshal profile response: %s", err)
	}

	return &result.Response, nil
}
