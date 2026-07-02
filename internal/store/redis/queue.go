package redis

import "github.com/hibiken/asynq"

type QueueService struct{ inspector *asynq.Inspector }

func NewQueueService(redisURL string) (*QueueService, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, err
	}
	return &QueueService{inspector: asynq.NewInspector(opts)}, nil
}
func (s *QueueService) GetQueueStats(queueName string) (*asynq.QueueInfo, error) {
	return s.inspector.GetQueueInfo(queueName)
}
func (s *QueueService) Close() {
	if s.inspector != nil {
		s.inspector.Close()
	}
}
