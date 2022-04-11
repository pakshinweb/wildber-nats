package main

import (
	"flag"
	stan "github.com/nats-io/stan.go"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	flag.Parse()
	args := flag.Args()

	sc, err := stan.Connect("test-cluster", "stan-pub")
	if err != nil {
		log.Fatalf("Can't connect: %v", err)
	}
	defer sc.Close()

	subj, msg := "wildber", args[0]

	filename, err := os.Open(msg)
	if err != nil {
		log.Fatal(err)
	}
	defer filename.Close()
	data, err := ioutil.ReadAll(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = sc.Publish(subj, data)
	log.Println(data)
	if err != nil {
		log.Fatalf("Error during publish: %v\n", err)
	}

}
