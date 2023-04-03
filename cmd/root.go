package cmd

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/marwatk/tstat-sensor-go/pkg/sensor"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func RootCmd() *cobra.Command {
	var logLevel string
	var cmd = &cobra.Command{
		Use:   "tstat-sensor-go",
		Short: "Simulate Wifi Temperature Sensor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("need subcommand")
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch logLevel {
			case "error":
				zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			case "warn":
				zerolog.SetGlobalLevel(zerolog.WarnLevel)
			case "info":
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			case "debug":
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			case "trace":
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			default:
				return fmt.Errorf("invalid log-level [%s]", logLevel)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (error,warn,info,debug,trace)")
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
	var celsius bool
	var pair bool
	var mac string
	var keyStr string
	var typeStr string
	var seqNum int
	var unitId int
	var addr string
	var cmd = &cobra.Command{
		Use:   "send [flags] -- <sensorName> <temperature>",
		Short: "Send a reading",
		Long: `Send a reading. Make sure to prefix your temperature by -- 
so that negative temps aren't treated as an errant flag`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			temp, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return fmt.Errorf("temperature not float: %w", err)
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

			return sensor.SimpleSend(sensor.Temperature{Value: temp, Celsius: celsius}, args[0], pair, mac, keyB, sensorType, seqNum, unitId, addr)
		},
	}

	cmd.Flags().StringVarP(&addr, "address", "a", "255.255.255.255", "Address to send to")
	cmd.Flags().BoolVarP(&celsius, "celsius", "c", false, "Temp is Celsius")
	cmd.Flags().BoolVarP(&pair, "pair", "p", false, "Send as a pairing message")
	cmd.Flags().StringVarP(&mac, "mac", "m", "", "MAC address of simulated sensor (blank will be generated from sensorName)")
	cmd.Flags().StringVarP(&keyStr, "key", "k", "", "Signature Key (blank will be generated from sensorName)")
	cmd.Flags().StringVarP(&typeStr, "type", "t", "remote", "Sensor type (outdoor, remote, supply, return)")
	cmd.Flags().IntVarP(&seqNum, "seqnum", "s", -1, "Reading sequence number (-1 means generate from time of day)")
	cmd.Flags().IntVarP(&unitId, "unitid", "u", 1, "Unit ID")

	return cmd
}
