package main

import (
	"flag"
	"log"
	"mipha/miphacore"
	"os"
)

var (
	templateFolder  = flag.String("templates", "", "folder contains template files")
	configFile      = flag.String("config", "", "file path to config file, only accept yaml format")
	helperTempalate = flag.String("helper", "", "flle path to customized helper template file, ex. helper.tpl")
	outputFolder    = flag.String("output", "", "folder to generate the output")
	help            = flag.Bool("help", false, "Print this help message")
	logger          = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
)

func main() {
	flag.Parse()
	if len(*templateFolder) == 0 ||
		len(*configFile) == 0 ||
		len(*outputFolder) == 0 ||
		*help {
		flag.PrintDefaults()
		return
	}

	logger.Printf(">> Parameters:")
	logger.Printf("templates: %s", *templateFolder)
	logger.Printf("helper   : %s", *helperTempalate)
	logger.Printf("config   : %s", *configFile)
	logger.Printf("output   : %s", *outputFolder)

	m := miphacore.NewMipha(
		*templateFolder,
		*configFile,
		*helperTempalate,
		*outputFolder,
	)

	if err := m.Load(); err != nil {
		log.Fatal(err)
	}

	if err := m.Execute(); err != nil {
		log.Fatal(err)
	}

	logger.Println("Process done!")

}
