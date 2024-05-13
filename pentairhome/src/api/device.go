package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type DeviceRequest struct {
	DeviceIds []string `json:"deviceIds"`
}

type ProductInfo struct {
	Visible  bool   `json:"visible"`
	NickName string `json:"nickName"`
	PoolID   any    `json:"poolId"`
	Maker    string `json:"maker"`
	Survey   any    `json:"survey"`
	Model    string `json:"model"`
	Order    int    `json:"order"`
}

type DeviceField struct {
	Name  string `json:"name"`
	Min   string `json:"min"`
	Max   string `json:"max"`
	Value string `json:"value"`
}

type Device struct {
	DeviceType   string                 `json:"deviceType"`
	DeviceID     string                 `json:"deviceId"`
	MVersion     int                    `json:"mVersion"`
	FwVersion    string                 `json:"fwVersion"`
	Timestamp    string                 `json:"timestamp"`
	Delivered    int64                  `json:"delivered"`
	DebOffStat   bool                   `json:"deb_off_stat"`
	DebOffTime   int64                  `json:"deb_off_time"`
	Alarm        bool                   `json:"alarm"`
	Online       bool                   `json:"online"`
	Fields       map[string]DeviceField `json:"fields"`
	ProductInfo  ProductInfo            `json:"productInfo"`
	Pname        string                 `json:"pname"`
	ReportedDate int64                  `json:"reportedDate"`
}

func (d Device) GetActualPower() (float64, error) {
	return strconv.ParseFloat(d.Fields["ifs3"].Value, 64)
}

func (d Device) GetActualSpeed() (float64, error) {
	return strconv.ParseFloat(d.Fields["ifs4"].Value, 64)
}

func (d Device) GetActualFlow() (float64, error) {
	return strconv.ParseFloat(d.Fields["ifs6"].Value, 64)
}

func (d Device) GetActualTemp() (float64, error) {
	return strconv.ParseFloat(d.Fields["t0"].Value, 64)
}

type DeviceResponse struct {
	Response struct {
		Data []Device `json:"data"`
		Code string   `json:"code"`
	} `json:"response"`
}

func (client APIClient) GetDevice(deviceId string) (*Device, error) {
	deviceRequest := DeviceRequest{
		DeviceIds: []string{deviceId},
	}

	jsonData, err := json.Marshal(deviceRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device request: %s", err)
	}

	body, err := client.MakeRequest("device2/device2-service/user/device", "POST", bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, fmt.Errorf("failed to get device: %s", err)
	}

	var result DeviceResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		return nil, fmt.Errorf("failed to unmarshal device response: %s", err)
	}

	if len(result.Response.Data) == 0 {
		return nil, fmt.Errorf("device not found: %s", deviceId)
	}

	return &result.Response.Data[0], nil
}
