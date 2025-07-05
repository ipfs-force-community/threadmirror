package xscraper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/michimani/gotwi/media/upload"
	"github.com/michimani/gotwi/media/upload/types"
	"github.com/michimani/gotwi/resources"
)

// uploadChunkSize is the fixed size for each upload chunk (2MB)
const uploadChunkSize = 2 * 1024 * 1024

// Pool for reusing chunk buffers to reduce memory allocation
var chunkBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, uploadChunkSize)
		return &buf // pointer-to-slice avoids SA6002 allocations
	},
}

// MediaUploadResult represents the result of a media upload operation
// Contains the media ID and key that can be used to attach the media to tweets
type MediaUploadResult struct {
	MediaID  string // Unique identifier for the uploaded media
	MediaKey string // Media key identifier for attachments
	Size     uint   // Size of the uploaded media in bytes
}

func (x *XScraper) UploadMedia(ctx context.Context, mediaReader io.Reader, mediaSize int) (*MediaUploadResult, error) {
	if x.gotwiClient == nil {
		return nil, fmt.Errorf("gotwi client not initialized. Call InitializeGotwiClient first")
	}

	if mediaSize <= 0 {
		return nil, fmt.Errorf("mediaSize must be greater than 0")
	}

	// Preserve original reader for upload while detecting MIME type.
	var (
		dataReader io.Reader
		mimeType   string
	)

	if rs, ok := mediaReader.(io.ReadSeeker); ok {
		// Detect and then rewind
		if m, err := mimetype.DetectReader(rs); err == nil {
			mimeType = m.String()
			_, _ = rs.Seek(0, io.SeekStart)
		}
		dataReader = rs
	} else {
		// Fallback: read entire bytes into memory
		buf, err := io.ReadAll(mediaReader)
		if err != nil {
			return nil, fmt.Errorf("read media into buffer: %w", err)
		}
		mimeType = mimetype.Detect(buf).String()
		dataReader = bytes.NewReader(buf)
	}

	// Step 1: Initialize upload
	initInput := &types.InitializeInput{
		TotalBytes: mediaSize,
		MediaType:  types.MediaType(mimeType),
	}

	initResult, err := upload.Initialize(ctx, x.gotwiClient, initInput)
	if err != nil {
		return nil, fmt.Errorf("initialize media upload: %w", err)
	}

	mediaID := initResult.Data.MediaID
	if mediaID == "" {
		return nil, fmt.Errorf("media ID not returned from initialize")
	}

	// Step 2: Upload media in chunks using streaming approach
	const chunkSize = uploadChunkSize

	// If media size is small enough, upload in a single segment to avoid unnecessary chunk logic
	if mediaSize <= chunkSize {
		appendInput := &types.AppendInput{
			MediaID:      mediaID,
			SegmentIndex: 0,
			Media:        dataReader,
		}

		_, err = upload.Append(ctx, x.gotwiClient, appendInput)
		if err != nil {
			return nil, fmt.Errorf("failed to append media: %w", err)
		}
	} else {
		// Chunked upload path
		totalChunks := (mediaSize + chunkSize - 1) / chunkSize

		// Get buffer from pool to avoid repeated allocations (pointer-to-slice to avoid SA6002)
		bufPtr := chunkBufferPool.Get().(*[]byte)
		defer chunkBufferPool.Put(bufPtr)
		chunkBuffer := *bufPtr

		reader := dataReader
		for i := range totalChunks {
			// Calculate expected read size for this chunk
			expectedSize := chunkSize
			if i == totalChunks-1 {
				// Last chunk might be smaller
				expectedSize = mediaSize - (i * chunkSize)
			}

			// Read chunk from stream into reusable buffer
			bytesRead := 0
			for bytesRead < expectedSize {
				n, readErr := reader.Read(chunkBuffer[bytesRead:expectedSize])
				bytesRead += n

				if readErr == io.EOF {
					if bytesRead == 0 {
						return nil, fmt.Errorf("unexpected EOF at chunk %d", i)
					}
					break
				}
				if readErr != nil {
					return nil, fmt.Errorf("read chunk %d: %w", i, readErr)
				}
			}

			// Create reader for the actual data read (avoid copying)
			chunkReader := bytes.NewReader(chunkBuffer[:bytesRead])

			appendInput := &types.AppendInput{
				MediaID:      mediaID,
				SegmentIndex: i,
				Media:        chunkReader,
			}

			_, err := upload.Append(ctx, x.gotwiClient, appendInput)
			if err != nil {
				return nil, fmt.Errorf("append media chunk %d: %w", i, err)
			}
		}
	}

	// Step 3: Finalize upload
	finalizeInput := &types.FinalizeInput{
		MediaID: mediaID,
	}

	finalizeResult, err := upload.Finalize(ctx, x.gotwiClient, finalizeInput)
	if err != nil {
		return nil, fmt.Errorf("finalize media upload: %w", err)
	}

	// Handle asynchronous processing if ProcessingInfo is present and not yet succeeded.
	processingInfo := finalizeResult.Data.ProcessingInfo
	for processingInfo.State == resources.ProcessingInfoStatePending || processingInfo.State == resources.ProcessingInfoStateInProgress {
		waitSecs := processingInfo.CheckAfterSecs
		if waitSecs <= 0 {
			waitSecs = 1
		}

		// Respect context cancellation while waiting.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(waitSecs) * time.Second):
		}

		statusInput := &types.FinalizeInput{MediaID: mediaID}
		statusResult, err := upload.Finalize(ctx, x.gotwiClient, statusInput)
		if err != nil {
			return nil, fmt.Errorf("check media processing status: %w", err)
		}

		finalizeResult = statusResult
		processingInfo = statusResult.Data.ProcessingInfo
	}

	if processingInfo.State == resources.ProcessingInfoStateFailed {
		return nil, fmt.Errorf("media processing failed")
	}

	result := &MediaUploadResult{
		MediaID:  finalizeResult.Data.MediaID,
		MediaKey: finalizeResult.Data.MediaKey,
		Size:     finalizeResult.Data.Size,
	}

	return result, nil
}
