package resumes

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher publishes resume generation jobs to RabbitMQ for the resume worker.
type RabbitMQPublisher interface {
	PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error
	Close() error
}

// ResumeWorkerJob is the message published to RabbitMQ for the resume worker.
type ResumeWorkerJob struct {
	JobID          string                 `json:"jobId"`
	UserID         string                 `json:"userId"`
	JobDescription string                 `json:"jobDescription"`
	Metadata       map[string]interface{} `json:"metadata"`
}

const (
	resumeExchange   = "woragis.tasks"
	resumeQueue      = "resumes.queue"
	resumeRoutingKey = "resumes.generate"
)

type rabbitMQPublisher struct {
	channel *amqp.Channel
	logger  *slog.Logger
}

// NewRabbitMQPublisher creates a new RabbitMQ publisher for resume jobs.
func NewRabbitMQPublisher(channel *amqp.Channel, logger *slog.Logger) (RabbitMQPublisher, error) {
	// Declare exchange
	err := channel.ExchangeDeclare(
		resumeExchange,  // name
		"direct",        // kind
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		resumeQueue, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		resumeQueue,      // queue name
		resumeRoutingKey, // routing key
		resumeExchange,   // exchange
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &rabbitMQPublisher{
		channel: channel,
		logger:  logger,
	}, nil
}

// PublishResumeGenerationJob publishes a resume generation job to RabbitMQ.
func (p *rabbitMQPublisher) PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error {
	if p.channel == nil {
		return fmt.Errorf("channel is not available")
	}

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		resumeExchange,   // exchange
		resumeRoutingKey, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		p.logger.Error("failed to publish job",
			slog.String("jobId", job.JobID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to publish job: %w", err)
	}

	p.logger.Info("resume generation job published",
		slog.String("jobId", job.JobID),
		slog.String("userId", job.UserID),
	)

	return nil
}

// Close closes the RabbitMQ channel.
func (p *rabbitMQPublisher) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}

// NoOpPublisher is a no-op RabbitMQ publisher for when RabbitMQ is not available
type noOpPublisher struct {
	logger *slog.Logger
}

// NewNoOpPublisher creates a new no-op RabbitMQ publisher
func NewNoOpPublisher(logger *slog.Logger) RabbitMQPublisher {
	return &noOpPublisher{logger: logger}
}

// PublishResumeGenerationJob is a no-op implementation
func (p *noOpPublisher) PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error {
	p.logger.Warn("RabbitMQ publisher is not available, job will not be queued",
		slog.String("jobId", job.JobID),
		slog.String("userId", job.UserID),
	)
	return nil
}

// Close is a no-op implementation
func (p *noOpPublisher) Close() error {
	return nil
}
