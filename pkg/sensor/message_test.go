package sensor

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestProto(t *testing.T) {
	keyStr := "NjBg/J+jAs9vLEbpxqCQyUg6l/drSD7DFd4MvRASCNs="
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		t.Fatalf("error deocding key")
	}
	hexString := "082ad2025a0a2a0800100f1a0c3230393134383235613965362001280930013801420753656e736f7231480350ff015861122c63356d435a4753574b797234734e355474524f6949476a4842654336316f7a6b5a6430636b5073716559773d"
	binary, err := hex.DecodeString(hexString)
	if err != nil {
		t.Fatalf("Error converting hex: %v", err)
	}
	msg := &SensorMsg{}
	err = proto.Unmarshal(binary, msg)
	if err != nil {
		t.Fatalf("Error unmarshalling: %v", err)
	}
	assert.Equal(t, "Sensor1", *msg.DataWithHash.SensorData.SensorName, "Sensor name matches")
	assert.NoError(t, ValidateSignature(msg, key), "Message validates")
}

func TestTempFToMsg(t *testing.T) {
	test := func(expected int32, tempF float64) {
		t.Run(fmt.Sprintf("%f", tempF), func(t *testing.T) {
			assert.Equal(t, &expected, Temperature{Value: tempF}.ToMsg())
		})
	}
	// Actual results from thermostat sending test values:
	test(200, 140)
	test(157, 101)
	test(156, 100)
	test(150, 95)
	test(120, 68)
	test(118, 66)
	test(117, 65)
	test(116, 64)
	test(114, 63)
	test(113, 62)
	test(112, 61)
	test(111, 60)
	test(110, 59)
	test(60, 14)
	test(2, -38)
	test(0, -40)
}
