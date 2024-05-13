package config

import (
	"os"
	"reflect"
	"testing"
)

func TestFetchRuntimeConfiguration(t *testing.T) {
	// Set up test flags
	os.Args = []string{
		"test",
		"--pentairhome_username=testuser",
		"--pentairhome_password=testpassword",
		"--mqtt_host=testhost",
		"--mqtt_port=testport",
		"--mqtt_username=testusername",
		"--mqtt_password=testpassword",
	}

	// Call the function
	config := FetchRuntimeConfiguration()

	// Assert the expected values
	expectedConfig := &RuntimeConfiguration{
		PentairHomeUsername: "testuser",
		PentairHomePassword: "testpassword",
		MQTTHost:            "testhost",
		MQTTPort:            "testport",
		MQTTUsername:        "testusername",
		MQTTPassword:        "testpassword",
	}

	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("FetchRuntimeConfiguration() = %v, want %v", config, expectedConfig)
	}
}

func TestFetchConfiguration(t *testing.T) {
	// Call the function
	config := FetchConfiguration()

	// Assert the expected values
	expectedConfig := &Configuration{
		AWSRegion:         "us-west-2",
		AWSUserPoolID:     "us-west-2_lbiduhSwD",
		AWSClientID:       "3de110o697faq7avdchtf07h4v",
		AWSIdentityPoolId: "us-west-2:6f950f85-af44-43d9-b690-a431f753e9aa",
	}

	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("FetchConfiguration() = %v, want %v", config, expectedConfig)
	}
}

func TestGetLoginKey(t *testing.T) {
	// Create a test configuration
	config := &Configuration{
		AWSRegion:         "us-west-2",
		AWSUserPoolID:     "us-west-2_lbiduhSwD",
		AWSClientID:       "3de110o697faq7avdchtf07h4v",
		AWSIdentityPoolId: "us-west-2:6f950f85-af44-43d9-b690-a431f753e9aa",
	}

	// Call the method
	loginKey := config.GetLoginKey()

	// Assert the expected value
	expectedLoginKey := "cognito-idp.us-west-2.amazonaws.com/us-west-2_lbiduhSwD"
	if loginKey != expectedLoginKey {
		t.Errorf("GetLoginKey() = %s, want %s", loginKey, expectedLoginKey)
	}
}

func getBaseConfig() *RuntimeConfiguration {
	return &RuntimeConfiguration{
		PentairHomeUsername: "PentairHomeUsername",
		PentairHomePassword: "PentairHomePassword",
		MQTTHost:            "MQTTHost",
		MQTTPort:            "MQTTPort",
		MQTTUsername:        "MQTTUsername",
		MQTTPassword:        "MQTTPassword",
	}
}

func TestValidateRuntimeConfiguration(t *testing.T) {
	fields := []string{"PentairHomeUsername", "PentairHomePassword", "MQTTHost", "MQTTPort", "MQTTUsername", "MQTTPassword"}

	for _, field := range fields {
		config := getBaseConfig()
		reflect.ValueOf(config).Elem().FieldByName(field).SetString("")
		errors := config.ValidateRuntimeConfiguration()

		if len(errors) != 1 {
			t.Errorf("ValidateRuntimeConfiguration() = %v, want 1 error", errors)
		}

		expectedError := field + " is required"
		if errors[0].Error() != expectedError {
			t.Errorf("ValidateRuntimeConfiguration() = %v, want %s", errors[0].Error(), expectedError)
		}
	}
}
