package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 寫檔工具的錯誤直接輸出至終端機，避免一般 logger 寫回自身造成遞迴或死鎖。
var diagnosticLogger = log.New(os.Stderr, "", log.LstdFlags)

type DailyWriter struct {
	mu       sync.Mutex
	dir      string
	prefix   string
	date     string
	file     *os.File
	closed   bool
	location *time.Location
}

// 建立以台灣日期分檔的日誌輸出，同一天追加寫入。
func NewDailyWriter(dir, prefix string) (*DailyWriter, error) {
	if prefix == "" || filepath.Base(prefix) != prefix {
		return nil, fmt.Errorf("日誌檔名前綴不正確")
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("建立日誌目錄失敗: %w", err)
	}
	writer := &DailyWriter{
		dir:      dir,
		prefix:   prefix,
		location: time.FixedZone("Asia/Taipei", 8*60*60),
	}
	if err := writer.rotate(time.Now()); err != nil {
		return nil, err
	}
	return writer, nil
}

// 每次寫入都檢查日期，跨日後的第一筆日誌自動切換檔案。
func (w *DailyWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return 0, os.ErrClosed
	}
	if err := w.rotate(time.Now()); err != nil {
		diagnosticLogger.Printf("[log] 每日日誌切換失敗: prefix=%s err=%v", w.prefix, err)
		return os.Stderr.Write(data)
	}
	n, err := w.file.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		diagnosticLogger.Printf("[log] 日誌寫入失敗: prefix=%s date=%s err=%v", w.prefix, w.date, err)
		remaining, fallbackErr := os.Stderr.Write(data[n:])
		if fallbackErr != nil {
			return n + remaining, fallbackErr
		}
		return n + remaining, nil
	}
	return n, nil
}

// 呼叫端須持有鎖，建構期間尚未共享時除外。
func (w *DailyWriter) rotate(now time.Time) error {
	date := now.In(w.location).Format("2006-01-02")
	if w.file != nil && w.date == date {
		return nil
	}
	path := filepath.Join(w.dir, w.prefix+"-"+date+".log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("開啟每日日誌失敗: path=%s: %w", path, err)
	}
	previous := w.file
	w.file = file
	w.date = date
	if previous != nil {
		if err := previous.Close(); err != nil {
			diagnosticLogger.Printf("[log] 關閉前一日日誌失敗: prefix=%s err=%v", w.prefix, err)
		}
	}
	return nil
}

// 關閉檔案並阻止後續寫入。
func (w *DailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	if w.file == nil {
		return nil
	}
	return w.file.Close()
}
