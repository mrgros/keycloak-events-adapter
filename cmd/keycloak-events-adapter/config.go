package main

// Config конфигурация приложения
type Config struct {
	LogLevel   string `long:"log-level" description:"Log level: panic, fatal, warn or warning, info, debug" env:"LOG_LEVEL" required:"true"`
	LogJSON    bool   `long:"log-json" description:"Enable force log format JSON" env:"LOG_JSON"`
	GrpcListen string `long:"grpc-listen" description:"Listening host:port for grpc-server" env:"GRPC_LISTEN" required:"true"`

	TntHost     string `long:"tnt-host" description:"Tarantool host" env:"TNT_HOST" required:"true"`
	TntPort     int    `long:"tnt-port" description:"Tarantool port" env:"TNT_PORT" required:"true"`
	TntUser     string `long:"tnt-user" description:"Tarantool user" env:"TNT_USER" required:"true"`
	TntPassword string `long:"tnt-password" description:"Tarantool password" env:"TNT_PASSWORD" required:"true"`
}
