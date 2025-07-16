package queue

import (
	"context"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestScreenshot(t *testing.T) {
	opts := util.DefaultScreenshotOptions()
	opts.AdditionalOptions = []chromedp.ExecAllocatorOption{chromedp.DisableGPU}

	buf, err := util.TakeScreenshotFromURL(context.Background(), "https://www.google.com", opts)
	require.NoError(t, err)
	require.NotEmpty(t, buf)
}
