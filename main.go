package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

const INFO = "INFO"
const WARNING = "WARNING"
const ERROR = "ERROR"
const TRACE = "TRACE"
const DEBUG = "DEBUG"

var connection *amqp.Connection
var channel *amqp.Channel
var queue amqp.Queue

var project_endpoint_type string
var project_code string

func InitLogger() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("failed get env for asiasiapac logger, error: %s", err)
	}

	user := os.Getenv("RABBITMQ_LOGGER_USERNAME")
	pass := os.Getenv("RABBITMQ_LOGGER_PASSWORD")
	host := os.Getenv("RABBITMQ_LOGGER_HOST")
	port := os.Getenv("RABBITMQ_LOGGER_PORT")

	project_endpoint_type = os.Getenv("PROJECT_ENDPOINT_TYPE")
	project_code = os.Getenv("PROJECT_CODE")

	connection, err = amqp.DialConfig(fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, pass, host, port, user), amqp.Config{
		Heartbeat: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to init connection asiasiapac logger, error: %s", err)
	}

	channel, err = connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to init channel asiasiapac logger, error: %s", err)
	}

	queue, err = channel.QueueDeclare(
		"logger-sys",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create queue asiasiapac logger, error: %s", err)
	}

	return nil
}

func Save(level, task, remark, note string) error {
	if channel != nil {
		message := map[string]string{
			"remark":         remark,
			"level":          level,
			"endpoint_rules": task,
			"note":           note,
			"endpoint_type":  project_endpoint_type,
			"source":         project_code,
			"time":           time.Now().Format("2006-01-02T15:04:05.000000000Z07:00"),
			"uuid":           uuidV4(),
			"refer_class":    "go function",
		}

		msg, err := json.Marshal(&message)
		if err != nil {
			return fmt.Errorf("failed to marshal message asiasiapac logger, error: %s", err)
		}

		err = channel.PublishWithContext(
			context.Background(),
			"",
			queue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        msg,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to publish message asiasiapac logger, error: %s", err)
		}
	}

	return nil
}

func uuidV4() string {
	return fmt.Sprintf("%04x%04x-%04x-%04x-%04x-%04x%04x%04x",
		rand.Intn(0xffff), rand.Intn(0xffff), // 32 bits for "time_low"
		rand.Intn(0xffff),                                       // 16 bits for "time_mid"
		rand.Intn(0x0fff)|0x4000,                                // 16 bits for "time_hi_and_version", with version number 4
		rand.Intn(0x3fff)|0x8000,                                // 16 bits for "clk_seq_hi_res" and "clk_seq_low"
		rand.Intn(0xffff), rand.Intn(0xffff), rand.Intn(0xffff), // 48 bits for "node"
	)
}
