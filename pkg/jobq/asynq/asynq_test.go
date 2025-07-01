package asynqjobq

import (
	"context"
	"os"
	"testing"
	"time"

	"log/slog"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
)

// JobHandlerFunc is a helper to turn a function into a JobHandler for testing.
type JobHandlerFunc func(ctx context.Context, job *jobq.Job) error

func (f JobHandlerFunc) HandleJob(ctx context.Context, job *jobq.Job) error {
	return f(ctx, job)
}

func TestAsynqClientServerWithMiniredis(t *testing.T) {
	s := miniredis.RunT(t)
	defer s.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   0,
	})

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := NewAsynqServer(redisClient, logger)
	client := NewAsynqClient(redisClient)

	jobType := "test_job"
	jobPayload := []byte(`{"foo":"bar"}`)
	jobHandled := make(chan *jobq.Job, 1)

	handler := JobHandlerFunc(func(ctx context.Context, job *jobq.Job) error {
		jobHandled <- job
		return nil
	})
	server.RegisterHandler(jobType, handler)

	go func() {
		require.NoError(t, server.Start())
	}()
	defer server.Server.Shutdown()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	id, err := client.Enqueue(context.Background(), &jobq.Job{Type: jobType, Payload: jobPayload})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	select {
	case job := <-jobHandled:
		require.Equal(t, jobType, job.Type)
		require.Equal(t, jobPayload, job.Payload)
	case <-time.After(2 * time.Second):
		t.Fatal("job handler was not called")
	}
}
