package log

import (
	"github.com/stretchr/testify/require"
	log_v1 "go_log/api/v1"
	"io"
	"os"
	"testing"
)

func TestSegment(t *testing.T) {
	dir, _ := os.MkdirTemp("", "segment_test")
	defer os.RemoveAll(dir)

	want := &log_v1.Record{Value: []byte("Log item 1")}

	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entWidth * 3

	s, err := newSegment(dir, 16, c)
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset)
	require.Falsef(t, s.IsMaxed(), "segment should not be maxed")

	for i := 0; i < 3; i++ {
		offset, err := s.Append(want)
		require.NoError(t, err)
		require.Equal(t, uint64(i+16), offset)

		got, err := s.Read(offset)
		require.NoError(t, err)
		require.Equal(t, want.Value, got.Value)
	}

	_, err = s.Append(want)
	require.Equal(t, io.EOF, err)

	require.Truef(t, s.IsMaxed(), "segment should be maxed")

	c.Segment.MaxStoreBytes = uint64(len(want.Value) * 3)
	c.Segment.MaxIndexBytes = 1024

	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)

	require.Truef(t, s.IsMaxed(), "segment should be maxed")

	err = s.Remove()
	require.NoError(t, err)
	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	require.Falsef(t, s.IsMaxed(), "segment should not be maxed")
}
