package sensor

import (
	"fmt"
	"pentairhome/api"
)

type DiscoveryDevice struct {
	Name         string   `json:"name"`
	Identifiers  []string `json:"identifiers"`
	Manufacturer string   `json:"manufacturer"`
	SwVersion    string   `json:"sw_version"`
}

type SensorConfig struct {
	Name              string          `json:"name"`
	StateTopic        string          `json:"state_topic"`
	DeviceClass       string          `json:"device_class"`
	ValueTemplate     string          `json:"value_template"`
	UniqueID          string          `json:"unique_id"`
	Device            DiscoveryDevice `json:"device"`
	UnitOfMeasurement string          `json:"unit_of_measurement"`
}

type SensorData struct {
	Power       float64 `json:"power"`
	ActualSpeed float64 `json:"actualspeed"`
	ActualFlow  float64 `json:"actualflow"`
	ActualTemp  float64 `json:"actualtemp"`
	OutsideTemp float64 `json:"outsidetemp"`
}

func GenerateSensorConfig(device *api.Device, sensorName, sensorID, deviceClass, unitOfMeasurement string) SensorConfig {
	return SensorConfig{
		Name:              sensorName,
		UniqueID:          fmt.Sprintf("ph_%s_%s", device.DeviceID, sensorID),
		StateTopic:        fmt.Sprintf("pentairhome/%s", device.DeviceID),
		DeviceClass:       deviceClass,
		ValueTemplate:     fmt.Sprintf("{{ value_json.%s }}", sensorID),
		UnitOfMeasurement: unitOfMeasurement,
		Device: DiscoveryDevice{
			Name:         device.ProductInfo.NickName,
			Identifiers:  []string{device.DeviceID},
			Manufacturer: device.ProductInfo.Maker,
			SwVersion:    device.FwVersion,
		},
	}
}
