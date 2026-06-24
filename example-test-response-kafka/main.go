package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	outputDir := "response"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	brokerList := strings.Split(os.Getenv("KAFKA_BROKERSURLS"), ",")

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer(brokerList, config)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Panic(err)
		}
	}()

	topics := []string{
		os.Getenv("KAFKA_TOPIC_QUEUESENDINGMESSAGES"),
		os.Getenv("KAFKA_TOPIC_SENDSWARMCONTACTGIO"),
		os.Getenv("KAFKA_TOPIC_DIALOGUESWARMGIO"),
		os.Getenv("KAFKA_TOPIC_DEVINFPREQUEST"),
		os.Getenv("KAFKA_TOPIC_DEVINFPRESPONSE"),
		os.Getenv("KAFKA_TOPIC_DEVINFPDLQ"),
	}

	var wg sync.WaitGroup

	for _, topic := range topics {
		if topic == "" {
			continue
		}

		partitionList, err := consumer.Partitions(topic)
		if err != nil {
			log.Printf("[%s] Erro ao buscar partições para %s: %v", time.Now().Format("15:04:05"), topic, err)
			continue
		}

		for _, partition := range partitionList {
			pc, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
			if err != nil {
				log.Printf("[%s] Erro ao iniciar consumidor: %v", time.Now().Format("15:04:05"), err)
				continue
			}

			wg.Add(1)
			go func(pc sarama.PartitionConsumer, t string) {
				defer wg.Done()
				for msg := range pc.Messages() {
					filename := filepath.Join(outputDir, fmt.Sprintf("%s.txt", t))

					f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						log.Printf("[%s] Erro ao abrir arquivo %s: %v", time.Now().Format("15:04:05"), filename, err)
						continue
					}

					if _, err := f.Write(append(msg.Value, '\n')); err != nil {
						log.Printf("[%s] Erro ao escrever no arquivo %s: %v", time.Now().Format("15:04:05"), filename, err)
					} else {
						fmt.Printf("[%s] Mensagem salva em: %s | Tópico: %s\n", time.Now().Format("15:04:05"), filename, t)
					}
					f.Close()
				}
			}(pc, topic)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	fmt.Println("Serviço iniciado. Aguardando mensagens...")

	run := true
	for run {
		select {
		case <-signals:
			fmt.Println("\nEncerrando aplicação...")
			run = false
		case t := <-ticker.C:
			fmt.Printf("[%s] Kafka Consumer rodando... Monitorando %d tópicos.\n", t.Format("15:04:05"), len(topics))
		}
	}
}
