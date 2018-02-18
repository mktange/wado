package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/mktange/wado/pkg/wado"
	"gopkg.in/yaml.v2"
)

var configFile = flag.String("config", "./wado.yml", "Path to wado.yml")

type config struct {
	Wados []wado.Config `yaml:"wados"`
}

func main() {
	flag.Parse()

	dir := path.Dir(*configFile)
	os.Chdir(dir)

	yamlFile, err := ioutil.ReadFile(path.Base(*configFile))
	if err != nil {
		panic(err)
	}

	var conf config
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		panic(err)
	}

	wados := []wado.Instance{}
	for _, wadoConfig := range conf.Wados {
		instance, err := wado.New(wadoConfig)
		if err != nil {
			panic(err)
		}
		wados = append(wados, instance)
	}

	shutdownOnSignal(wados)
	select {}
}

func shutdownOnSignal(instances []wado.Instance) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		log.Println("[Wado] Got signal: ", s)

		wg := sync.WaitGroup{}
		for _, instance := range instances {
			wg.Add(1)
			go func() {
				instance.Kill()
				wg.Done()
			}()
		}
		wg.Wait()
		os.Exit(1)
	}()
}
