package logwriter

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type bufferCloseAction struct{}
type bufferWriteAction struct {
	Data string
}

type bufferTestWriter struct {
	Actions []interface{}
}

func (b *bufferTestWriter) Write(p []byte) (n int, err error) {
	b.Actions = append(b.Actions, &bufferWriteAction{Data: string(p)})
	return len(p), nil
}

func (b *bufferTestWriter) Close() error {
	b.Actions = append(b.Actions, &bufferCloseAction{})
	return nil
}

func TestBuffer_Write(t *testing.T) {
	type WriteOperation struct {
		data string
		err  error
	}
	type CloseOperation struct {
		err error
	}
	type UpdateDurationOperation struct {
		elapsed time.Duration
	}

	cases := []struct {
		name     string
		size     int
		interval time.Duration
		// List of WriteOperation and CloseOperation and UpdateDurationOperation().
		ops []interface{}
		// List of bufferWriteAction and bufferCloseAction.
		actions []interface{}
	}{
		{
			name: "close without write",
			ops: []interface{}{
				&CloseOperation{},
			},
			actions: []interface{}{
				&bufferCloseAction{},
			},
		}, {
			name: "write small data once and close",
			size: 100,
			ops: []interface{}{
				&WriteOperation{data: "small data"},
				&CloseOperation{},
			},
			actions: []interface{}{
				&bufferWriteAction{Data: "small data"},
				&bufferCloseAction{},
			},
		}, {
			name: "write large data once and close",
			size: 10,
			ops: []interface{}{
				&WriteOperation{data: "large data"},
				&CloseOperation{},
			},
			actions: []interface{}{
				&bufferWriteAction{Data: "large data"},
				&bufferCloseAction{},
			},
		}, {
			name: "write small data multiple times and close",
			size: 20,
			ops: []interface{}{
				&WriteOperation{data: "write "},    // +6 byte, total 6 byte.
				&WriteOperation{data: "small "},    // +6 byte, total 12 byte.
				&WriteOperation{data: "data "},     // +5 byte, total 17 byte.
				&WriteOperation{data: "multiple "}, // +9 byte, total 26 byte, flush buffer automatically.
				&WriteOperation{data: "times.\n"},  // +7 byte, total 7 byte.
				&CloseOperation{},
			},
			actions: []interface{}{
				&bufferWriteAction{Data: "write small data multiple "},
				&bufferWriteAction{Data: "times.\n"},
				&bufferCloseAction{},
			},
		}, {
			name: "trigger flush() automatically when interval passed",
			size: 10000,
			ops: []interface{}{
				&WriteOperation{data: "small data\n"},
				&UpdateDurationOperation{elapsed: 1500 * time.Millisecond},
				&WriteOperation{data: "should be flush after write this message."},
				&UpdateDurationOperation{elapsed: 1600 * time.Millisecond},
				&WriteOperation{data: "next message\n"},
				&UpdateDurationOperation{elapsed: 2100 * time.Millisecond}, // 2100 ms passed, but flush() is not called because only 600ms have passed since the last flush() called.
				&WriteOperation{data: "next message 2\n"},
				&UpdateDurationOperation{elapsed: 2600 * time.Millisecond},
				&WriteOperation{data: "should be flush after write this message."},
				&WriteOperation{data: "next message 3\n"},
				&CloseOperation{},
			},
			actions: []interface{}{
				&bufferWriteAction{Data: "small data\nshould be flush after write this message."},
				&bufferWriteAction{Data: "next message\nnext message 2\nshould be flush after write this message."},
				&bufferWriteAction{Data: "next message 3\n"},
				&bufferCloseAction{},
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			currentTime := time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)
			elapsed := time.Duration(0)
			w := &bufferTestWriter{}
			now := func() time.Time {
				return currentTime.Add(elapsed)
			}
			buf := newBuffer(testCase.size, 1*time.Second, w, now)

			for _, _op := range testCase.ops {
				switch op := _op.(type) {
				case *WriteOperation:
					n, err := buf.Write([]byte(op.data))
					assert.Equal(t, op.err, err)
					assert.Equal(t, len(op.data), n)
				case *CloseOperation:
					err := buf.Close()
					assert.Equal(t, op.err, err)
				case *UpdateDurationOperation:
					elapsed = op.elapsed
				default:
					assert.Failf(t, "unsupported type", "%#+v", _op)
				}
			}

			assert.Equal(t, testCase.actions, w.Actions)
		})
	}
}
