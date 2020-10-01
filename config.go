package homekit

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

type configStorage interface {
	Set(key, data []byte) error
	Get(key []byte) ([]byte, error)
}

type Config struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	ServePort    int    `json:"server_port"`
	Pin          string `json:"pin"`
	Version      int    `json:"version"`
	CategoryID   uint8  `json:"category_id"`
	State        int    `json:"state"`
	Protocol     string `json:"protocol"`
	Discoverable bool   `json:"discoverable"`
	MfiCompliant bool   `json:"mfi_compliant"`
	ConfigHash   []byte `json:"config_hash"`
	SetupID      string `json:"setup_id"`
	storage      configStorage
}

func NewConfig(name string, storage configStorage) *Config {
	return &Config{
		ID:           randomMac().String(),
		ServePort:    25736,
		Pin:          "00102003",
		Version:      1,
		Name:         name,
		CategoryID:   0,
		State:        1,
		Protocol:     "1.0",
		Discoverable: true,
		MfiCompliant: false,
		SetupID:      "HOMEKIT",
		storage:      storage,
	}
}

func (cfg Config) MDNSRecords() map[string]string {
	records := map[string]string{
		"pv": cfg.Protocol,
		"id": cfg.ID,
		"c#": strconv.Itoa(cfg.Version),
		"s#": strconv.Itoa(cfg.State),
		"sf": "0",
		"ff": "0",
		"md": cfg.Name,
		"ci": strconv.Itoa(int(cfg.CategoryID)),
		"sh": cfg.setupHash(),
	}
	if cfg.Discoverable {
		records["sf"] = "1"
	}

	if cfg.MfiCompliant {
		records["ff"] = "1"
	}

	return records
}

func (cfg *Config) Save() error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return cfg.storage.Set([]byte("config"), data)
}

func (cfg *Config) Load() error {
	data, err := cfg.storage.Get([]byte("config"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return json.Unmarshal(data, cfg)
}

func (cfg *Config) updateConfigHash(hash []byte) {
	if cfg.ConfigHash != nil && !bytes.Equal(hash, cfg.ConfigHash) {
		cfg.Version++
	}

	cfg.ConfigHash = hash
}

func (cfg *Config) setupHash() string {
	sum := sha512.Sum512([]byte(cfg.SetupID + cfg.ID))
	return base64.StdEncoding.EncodeToString([]byte{sum[0], sum[1], sum[2], sum[3]})
}

func randomMac() net.HardwareAddr {
	buf := make([]byte, 6)
	var mac net.HardwareAddr

	_, _ = rand.Read(buf)

	// Set the local bit
	buf[0] |= 2

	return append(mac, buf...)
}
