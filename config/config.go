package config

type Config interface {
	Key() string
	OnUpdate(key string, val string)
}
