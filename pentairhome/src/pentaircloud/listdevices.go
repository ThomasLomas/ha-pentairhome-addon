package pentaircloud

import (
	"encoding/json"
	"fmt"
)

type ListDevice struct {
	CreatedDate int64       `json:"createdDate"`
	DeviceType  string      `json:"deviceType"`
	AddressID   string      `json:"addressId"`
	Status      string      `json:"status"`
	Pname       string      `json:"pname"`
	Order       int         `json:"order"`
	DeviceID    string      `json:"deviceId"`
	ProductInfo ProductInfo `json:"productInfo"`
}

func (ld ListDevice) IsIntelliConnect() bool {
	return ld.ProductInfo.Model == "IntelliConnect"
}

type ListDevicesResponse struct {
	Response        []ListDevice `json:"response"`
	AllDevicesCount int          `json:"allDevicesCount"`
	Msgs            []any        `json:"msgs"`
}

func (client APIClient) ListDevices() ([]ListDevice, error) {
	body, bodyErr := client.MakeRequest("device2/device2-service/user/listdevices", "GET", nil)

	if bodyErr != nil {
		return nil, fmt.Errorf("failed to list devices: %s", bodyErr)
	}

	var result ListDevicesResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		return nil, fmt.Errorf("failed to unmarshal list devices response: %s", err)
	}

	return result.Response, nil
}
