package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/VoR0220/SimpleABI/generation"
	"github.com/VoR0220/SimpleABI/parser"

	"github.com/spf13/cobra"
)

var (
	abiFilename string
	encode      bool
	decode      bool
	language    string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&abiFilename, "abi", "a", "", "path of simpleabi file; must be in .abi extension; see docs for details")
	rootCmd.PersistentFlags().BoolVarP(&encode, "encode", "e", false, "enabling this flag generates an encoding abi template")
	rootCmd.PersistentFlags().BoolVarP(&decode, "decode", "d", false, "enabling this flag generates a decoding abi template")
	rootCmd.PersistentFlags().StringVarP(&language, "lang", "l", "c", "defines which language you would like to generate in, must be one of: c")
}

var rootCmd = &cobra.Command{
	Use:   "simpleabi",
	Short: "SimpleAbi is a tool for creating non solidity smart contracts for Qtum",
	Long: `SimpleAbi is a tool that takes in an input file specifically crafted for ABIs (see documentation
for how to make this properly work) and generates a template for smart contract interaction in a variety of available languages. 
Current languages available are C but we are adamently working hard at Qtum to add more in.`,
	Run: func(cmd *cobra.Command, args []string) {
		if encode == false && decode == false {
			fmt.Printf("Must select one of encode or decode (or both) as an option to use this tool\n")
			os.Exit(1)
		}

		if _, err := os.Stat(abiFilename); os.IsNotExist(err) {
			fmt.Printf("Please include a valid path to a valid .abi file\n")
			os.Exit(1)
		}

		if extension := filepath.Ext(abiFilename); extension != ".abi" {
			fmt.Printf("Expected file extension .abi, got %v\n", filepath.Ext(abiFilename))
			os.Exit(1)
		}

		nameBase := strings.TrimSuffix(filepath.Base(abiFilename), ".abi")

		if language != "c" {
			fmt.Printf("Unexpected language %v selected, select one of: c\n", language)
			os.Exit(1)
		}

		interfaceBuilder, err := parser.Parse(abiFilename)
		if err != nil {
			fmt.Printf("Error in parsing your abi file: %v\n", err)
			os.Exit(1)
		}

		if encode {
			name := nameBase + "ABI.c"
			var buf bytes.Buffer
			err := generation.GenerateTemplate(interfaceBuilder, nameBase+"ABI.c", &buf, true)
			if err != nil {
				fmt.Printf("Error in encoding template generation: %v\n", err)
			}
			err = ioutil.WriteFile(name, buf.Bytes(), 0666)
			if err != nil {
				fmt.Printf("Error in file creation and writing: %v\n", err)
			}
		}

		if decode {
			name := nameBase + "Dispatcher.c"
			var buf bytes.Buffer
			err := generation.GenerateTemplate(interfaceBuilder, nameBase+"Dispatcher.c", &buf, false)
			if err != nil {
				fmt.Printf("Error in encoding template generation: %v\n", err)
			}
			err = ioutil.WriteFile(name, buf.Bytes(), 0666)
			if err != nil {
				fmt.Printf("Error in file creation and writing: %v\n", err)
			}
		}

	},
}

// Execute runs the root command of the application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
