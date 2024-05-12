package api

import (
	"encoding/json"
	"log"
)

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
		log.Fatalf("failed to unmarshal profile response: %s", err)
	}

	return result.Response
}
