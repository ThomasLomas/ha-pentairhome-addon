package api

import (
	"bytes"
	"encoding/json"
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

func (d Device) GetActualPower() float64 {
	actualPower, _ := strconv.ParseFloat(d.Fields["ifs3"].Value, 64)
	return actualPower
}

func (d Device) GetActualSpeed() float64 {
	actualSpeed, _ := strconv.ParseFloat(d.Fields["ifs4"].Value, 64)
	return actualSpeed
}

func (d Device) GetActualFlow() float64 {
	actualFlow, _ := strconv.ParseFloat(d.Fields["ifs6"].Value, 64)
	return actualFlow
}

func (d Device) GetActualTemp() float64 {
	actualTemp, _ := strconv.ParseFloat(d.Fields["t0"].Value, 64)
	return actualTemp
}

type DeviceResponse struct {
	Response struct {
		Data []Device `json:"data"`
		Code string   `json:"code"`
	} `json:"response"`
}

func (client APIClient) GetDevice(deviceId string) Device {
	deviceRequest := DeviceRequest{
		DeviceIds: []string{deviceId},
	}

	jsonData, _ := json.Marshal(deviceRequest)
	body := client.MakeRequest("device2/device2-service/user/device", "POST", bytes.NewBuffer(jsonData))

	// log the response body
	// fmt.Println(string(body))

	var result DeviceResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		panic("Can not unmarshal DeviceResponse JSON")
	}

	return result.Response.Data[0]
}
