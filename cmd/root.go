/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/marwatk/tstat-sensor-go/pkg/sensor"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func RootCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "tstat-sensor-go",
		Short: "Simulate Wifi Temperature Sensor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("need subcommand")
		},
	}

	cmd.AddCommand(SendCmd())
	cmd.AddCommand(DumpCmd())
	return cmd
}

func DumpCmd() *cobra.Command {
	dupes := false
	last := ""
	var cmd = &cobra.Command{
		Use:   "dump",
		Short: "Listen and output messages as they arrive",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := net.ListenPacket("udp4", ":5001")
			if err != nil {
				return fmt.Errorf("error listening on port 5001: %w", err)
			}
			defer l.Close()

			buf := make([]byte, 2048)
			for {
				size, addr, err := l.ReadFrom(buf)
				if err != nil {
					return fmt.Errorf("error reading from socket: %w", err)
				}
				msg := sensor.SensorMsg{}
				err = proto.Unmarshal(buf[:size], &msg)
				if err != nil {
					fmt.Printf("error unmarshalling: %v\n", err)
				}
				this := msg.String()
				if dupes || this != last {
					fmt.Printf("From %s\n", addr)
					sensor.DumpMessage(&msg)
					last = this
					fmt.Println("")
				}
			}
		},
	}
	cmd.Flags().BoolVarP(&dupes, "show-duplicates", "d", false, "Show duplicate messages")

	return cmd
}

func SendCmd() *cobra.Command {
	var pair bool
	var mac string
	var keyStr string
	var typeStr string
	var seqNum int
	var unitId int
	var cmd = &cobra.Command{
		Use:   "send <sensorName> <tempF>",
		Short: "Send a reading",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			temp, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("temperature not integer: %w", err)
			}
			var keyB []byte
			if keyStr != "" {
				keyB = []byte(keyStr)
			}

			sensorType := sensor.SensorType_REMOTE
			switch strings.ToLower(typeStr) {
			case "outdoor":
				sensorType = sensor.SensorType_OUTDOOR
			case "remote":
				sensorType = sensor.SensorType_REMOTE
			case "supply":
				sensorType = sensor.SensorType_SUPPLY
			case "return":
				sensorType = sensor.SensorType_RETURN
			default:
				return fmt.Errorf("invalid sensor type [%s]", typeStr)
			}

			return sensor.SimpleSend(temp, args[0], pair, mac, keyB, sensorType, seqNum, unitId)
		},
	}

	cmd.Flags().BoolVarP(&pair, "pair", "p", false, "Send as a pairing message")
	cmd.Flags().StringVarP(&mac, "mac", "m", "", "MAC address of simulated sensor (blank will be generated from sensorName)")
	cmd.Flags().StringVarP(&keyStr, "key", "k", "", "Signature Key (blank will be generated from sensorName)")
	cmd.Flags().StringVarP(&typeStr, "type", "t", "remote", "Sensor type (outdoor, remote, supply, return)")
	cmd.Flags().IntVarP(&seqNum, "seqnum", "s", -1, "Reading sequence number (-1 means generate from time of day)")
	cmd.Flags().IntVarP(&unitId, "unitid", "u", 1, "Unit ID")

	return cmd
}
