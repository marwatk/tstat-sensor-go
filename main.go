/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"fmt"
	"os"

	"github.com/marwatk/tstat-sensor-go/cmd"
)

func main() {
	err := cmd.RootCmd().Execute()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error running cmd: %v", err))
	}
}
