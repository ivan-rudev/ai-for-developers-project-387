// Package config читает и валидирует конфигурацию приложения из YAML-файла.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config — корневая структура конфигурации (config.yaml).
type Config struct {
	Server    Server    `yaml:"server"`
	Health    Health    `yaml:"health"`
	RateLimit RateLimit `yaml:"rate_limit"`
	Admin     Admin     `yaml:"admin"`
	Default   Default   `yaml:"default"`
	Seed      Seed      `yaml:"seed"`
}

// Server — настройки HTTP-сервера.
type Server struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// Health — настройки healthcheck-эндпоинта.
type Health struct {
	Path string `yaml:"path"`
}

// RateLimit — настройки in-memory rate limiter.
type RateLimit struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

// Admin — привязка админской панели к владельцу.
type Admin struct {
	OwnerUUID string `yaml:"owner_uuid"`
}

// Default — дефолтные настройки календаря и события для новых владельцев.
type Default struct {
	WorkStart   string        `yaml:"work_start"`
	WorkEnd     string        `yaml:"work_end"`
	Timezone    string        `yaml:"timezone"`
	WorkingDays []string      `yaml:"working_days"`
	Events      []EventConfig `yaml:"events"`
}

// EventConfig — описание события по умолчанию.
type EventConfig struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	DurationMinutes int    `yaml:"duration_minutes"`
}

// Seed — тестовые владельцы и события, создаваемые при старте приложения.
type Seed struct {
	Owners []SeedOwner `yaml:"owners"`
	Events []SeedEvent `yaml:"events"`
}

// SeedOwner — тестовый владелец.
type SeedOwner struct {
	UUID  string `yaml:"uuid"`
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

// SeedEvent — тестовое событие.
type SeedEvent struct {
	OwnerUUID       string `yaml:"owner_uuid"`
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	DurationMinutes int    `yaml:"duration_minutes"`
}

// DefaultConfig возвращает конфигурацию с дефолтными значениями.
// Используется как базис перед чтением файла: отсутствующие секции
// подставляют безопасные значения.
func DefaultConfig() *Config {
	return &Config{
		Server: Server{Port: 8080, Host: "0.0.0.0"},
		Health: Health{Path: "/healthz"},
		RateLimit: RateLimit{
			RequestsPerMinute: 30,
			Burst:             10,
		},
		Default: Default{
			WorkStart:   "09:00",
			WorkEnd:     "18:00",
			Timezone:    "Europe/Moscow",
			WorkingDays: []string{"mon", "tue", "wed", "thu", "fri"},
		},
	}
}

// Load читает YAML-файл по пути path и возвращает конфигурацию.
// Значения, не указанные в файле, наследуются из DefaultConfig.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, nil
}
