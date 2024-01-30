package storage

import (
	"bufio"
	"io"
	"os"
	"sync"
)

type BackupStorage struct {
	mutex      sync.Mutex
	Gauges     map[string]float64 `json:"gauges"`
	Counters   map[string]int64   `json:"counters"`
	backupFile *os.File
	scanner    *bufio.Scanner
}

func NewBackupStorage(filename string) (*BackupStorage, error) {

	// открываем и очищаем
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &BackupStorage{
		backupFile: file,
		scanner:    bufio.NewScanner(file),
	}, nil
}

// сохранение в файл
func (b *BackupStorage) Save(dump string) error {

	b.mutex.Lock()
	defer b.mutex.Unlock()

	_, err := b.backupFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = b.backupFile.Truncate(0)
	if err != nil {
		return err
	}

	_, err = b.backupFile.Write([]byte(dump))
	if err != nil {
		return err
	}

	return nil
}

// получение данных из файла
func (b *BackupStorage) Load() (string, error) {

	b.mutex.Lock()
	defer b.mutex.Unlock()

	data, err := io.ReadAll(b.backupFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
