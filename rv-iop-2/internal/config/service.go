package config

type Service struct {
	App struct {
		Name               string `mapstructure:"name"`
		Environment        string `mapstructure:"env"`
		Port               int    `mapstructure:"port"`
		LogLevel           string `mapstructure:"log-level"`
		MetricSamplingRate string `mapstructure:"metric-sampling-rate"`
	} `mapstructure:"app"`
	ExternalService struct {
		Ab struct {
			Host                string `mapstructure:"host"`
			Port                int    `mapstructure:"port"`
			Auth                string `mapstructure:"auth"`
			ReadTimeOutInMs     int    `mapstructure:"read-timeout-in-ms"`
			DialTimeoutInMs     int    `mapstructure:"dial-timeout-in-ms"`
			MaxIdleConns        int    `mapstructure:"max-idle-conns"`
			MaxIdleConnsPerHost int    `mapstructure:"max-idle-conns-per-host"`
			IdleConnTimeoutInMs int    `mapstructure:"idle-conn-timeout-in-ms"`
		} `mapstructure:"ab"`
	} `mapstructure:"external-service"`
}
