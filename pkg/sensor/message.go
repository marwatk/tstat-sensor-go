package sensor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

var keyData = make(map[string][]byte)

func DumpMessage(msg *SensorMsg) {
	sigStatus := ""
	if *msg.Type == MessageType_PAIR {
		key, err := GetHashBytes(msg)
		if err != nil {
			sigStatus = fmt.Sprintf("Error decoding hash: %v", err)
		} else {
			sigStatus = "Pairing message (key received)"
			keyData[*msg.DataWithHash.SensorData.Mac] = key
		}
	} else {
		key, ok := keyData[*msg.DataWithHash.SensorData.Mac]
		if ok {
			err := ValidateSignature(msg, key)
			if err != nil {
				sigStatus = fmt.Sprintf("%v", err)
			} else {
				sigStatus = "Valid signature"
			}
		} else {
			sigStatus = "No key seen, press pair button on device to receive key data"
		}
	}
	fmt.Printf("Signature: %s\n", sigStatus)
	fmt.Println(msg.String())
}

func GetHashBytes(msg *SensorMsg) ([]byte, error) {
	return base64.StdEncoding.DecodeString(*msg.DataWithHash.Hash)
}

func ValidateSignature(msg *SensorMsg, key []byte) error {
	if msg.GetType() == MessageType_PAIR {
		return errors.New("can't validate a pairing message, hash is key not signature")
	}
	sig, err := GetHashBytes(msg)
	if err != nil {
		return fmt.Errorf("error decoding message signature")
	}
	sum, err := CalculateSignature(msg, key)
	if err != nil {
		return err
	}
	if !bytes.Equal(sum, sig) {
		return fmt.Errorf("signature not a match")
	}
	return nil
}

func CalculateSignature(msg *SensorMsg, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	data, err := proto.Marshal(msg.DataWithHash.SensorData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling submsg: %w", err)
	}
	h.Write(data)
	return h.Sum(nil), nil
}

func Send(msg *SensorMsg) error {
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:5001")
	if err != nil {
		return fmt.Errorf("error resolving broadcast address: %w", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("error dialing broadcast address: %w", err)
	}
	defer conn.Close()

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error writing packet: %w", err)
	}
	return nil
}

func FarenheitToMsg(tempF int) int {
	// TODO
	return tempF
}

func GenerateMAC(sensorName string) string {
	hash := sha256.Sum256([]byte(sensorName))
	// https://en.wikipedia.org/wiki/MAC_address#Ranges_of_group_and_locally_administered_addresses
	r := fmt.Sprintf("0a%x%x%x%x%x", hash[0], hash[1], hash[2], hash[3], hash[4])
	return r
}

func GenerateKey(sensorName string) []byte {
	r := sha256.Sum256([]byte(sensorName))
	return r[:]
}

// Real sensors increment this for each message, but that requires state so we fake
// it with time of day.
func GenerateSeqNum() int {
	now := time.Now().UTC()
	year, month, day := now.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	// Quarter-minute resolution
	return int(now.Sub(start).Seconds()) / 15
}

// BuildAndSend is a simple interface to send temp data. If pair is set
// it's sent as a pairing message, otherwise a normal data packet. If mac is empty it is generated
// from the sensorName. If key is nil it is generated fro the sensorName. If seqNum is -1 it is
// generated based on time of day. If sensorType is nil REMOTE is assumed..
func SimpleSend(tempF int, sensorName string, pair bool, mac string, key []byte, sensorType SensorType, seqNum int, unitId int) error {
	if unitId < 0 || unitId > 19 {
		return fmt.Errorf("unitId [%d] out of range (0-19)", unitId)
	}
	var seqNumP *int32
	if seqNum == -1 {
		seqNumP = intPointer(GenerateSeqNum())
	} else {
		seqNumP = intPointer(seqNum)
	}
	if mac == "" {
		mac = GenerateMAC(sensorName)
	}
	if key == nil {
		key = GenerateKey(sensorName)
	}
	unitIdP := intPointer(unitId)

	msg := &SensorMsg{
		DataWithHash: &DataWithHash{
			SensorData: &SensorData{
				UnitId:     unitIdP,
				Mac:        &mac,
				SensorType: &sensorType,
				Battery:    intPointer(95),
				Temp:       intPointer(int(FarenheitToMsg(tempF))),
				SensorName: &sensorName,
				SeqNum:     seqNumP,
			},
		},
	}
	SetUnknowns(msg)
	if pair {
		msg.Type = messageTypePointer(MessageType_PAIR)
		keyStr := base64.StdEncoding.EncodeToString(key)
		msg.DataWithHash.Hash = &keyStr
	} else {
		msg.Type = messageTypePointer(MessageType_DATA)
		sig, err := CalculateSignature(msg, key)
		if err != nil {
			return err
		}
		sigStr := base64.StdEncoding.EncodeToString(sig)
		msg.DataWithHash.Hash = &sigStr
	}
	return Send(msg)
}

func SetUnknowns(msg *SensorMsg) {
	// Don't know what these are and haven't seen them change
	msg.DataWithHash.SensorData.Field4 = intPointer(1)
	msg.DataWithHash.SensorData.Field5 = intPointer(9)
	msg.DataWithHash.SensorData.Field6 = intPointer(1)
	msg.DataWithHash.SensorData.Field7 = intPointer(1)
}

func intPointer(i int) *int32 {
	i32 := int32(i)
	return &i32
}

func sensorTypePointer(sensorType SensorType) *SensorType {
	return &sensorType
}

func messageTypePointer(msgType MessageType) *MessageType {
	return &msgType
}
