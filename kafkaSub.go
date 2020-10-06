package main

import (
	"context"
	"fmt"
	"gopkg.in/avro.v0"
	"os"

	"github.com/segmentio/kafka-go"
	"log"
)

type KeyValueObjects struct {
	message  string `json:"message"`
}

func getKafkaReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost://0.0.0.0:4222"},
		Topic: "reLoad",
	})
}

var reader *kafka.Reader

func kafkasubscriber() (*kafka.Reader){

	// connect to kafka
	reader = getKafkaReader()
	
	fmt.Println("start consuming ... !!")
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("message at topic:%v partition:%v offset:%v	%s = %s\n", m.Topic, m.Partition, m.Offset,string(m.Value))

		//write data into .avro file

		f, err := os.Create("./myfilefromKafka.avro")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = f.Write(m.Value)
		if err != nil {
			fmt.Println(err)

			return nil
		}
		recordReader()
	}
	
	return reader
	
	
	
}

func recordReader() {
	// Provide a filename to read and a DatumReader to manage the reading itself
	genericReader, err := avro.NewDataFileReader("myfilefromKafka.avro", avro.NewGenericDatumReader())
	if err != nil {
		// Should not actually happen
		panic(err)
	}

	for {
		obj := new(avro.GenericRecord)
		ok, err := genericReader.Next(obj)
		if !ok {
			if err != nil {
				panic(err)
			}
			break
		} else {
			values := obj.Map()
			fmt.Println("Message info : ====>", values["message"])

		}
	}
}
