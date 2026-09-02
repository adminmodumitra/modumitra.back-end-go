package obs

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

func TestOpenSinksDefaultStderr(t *testing.T) {
	for _, in := range [][]string{nil, {}, {""}} {
		w, err := OpenSinks(in)
		if err != nil {
			t.Fatalf("in=%v err=%v", in, err)
		}
		if w != os.Stderr {
			t.Fatalf("in=%v want stderr got %T", in, w)
		}
	}
	w, err := OpenSinks([]string{utils.SinkStderr})
	if err != nil {
		t.Fatal(err)
	}
	if w != os.Stderr {
		t.Fatalf("stderr sink: %T", w)
	}
	w, err = OpenSinks([]string{utils.SinkStdout})
	if err != nil {
		t.Fatal(err)
	}
	if w != os.Stdout {
		t.Fatalf("stdout sink: %T", w)
	}
}

func TestOpenSinksMultiWriter(t *testing.T) {
	w, err := OpenSinks([]string{utils.SinkStderr, utils.SinkStdout})
	if err != nil {
		t.Fatal(err)
	}
	if w == os.Stderr || w == os.Stdout {
		t.Fatal("expected MultiWriter, got a single std stream")
	}
}

func TestOpenSinksLokiReserved(t *testing.T) {
	_, err := OpenSinks([]string{utils.SinkLoki})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), `log sink "loki" is reserved`) {
		t.Fatalf("err=%v", err)
	}
}

func TestOpenSinksCloudWatchReserved(t *testing.T) {
	_, err := OpenSinks([]string{utils.SinkCloudWatch})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), `log sink "cloudwatch" is reserved`) {
		t.Fatalf("err=%v", err)
	}
}

func TestOpenSinksUnknown(t *testing.T) {
	_, err := OpenSinks([]string{"foo"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unknown log sink "foo"`) {
		t.Fatalf("err=%v", err)
	}
}

func TestInitLoggerWithWriterJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	prev := Logger()
	InitLoggerWithWriter(buf, "info", "cabinet-test")
	t.Cleanup(func() { SetLogger(prev) })
	l := Logger()
	l.Info().Msg("hello-sink")
	raw := strings.TrimSpace(buf.String())
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("json: %v raw=%s", err, raw)
	}
	if m["message"] != "hello-sink" {
		t.Fatalf("line=%v", m)
	}
	if m["service"] != "cabinet-test" {
		t.Fatalf("service=%v", m["service"])
	}
}
