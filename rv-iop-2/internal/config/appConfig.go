package config

type AppConfig struct {
	ZkConf      Zk
	ServiceConf *Service
}

func (c *AppConfig) GetStaticConfig() interface{} {
	return &c.ServiceConf
}

func (c *AppConfig) GetDynamicConfig() interface{} {
	return &c.ZkConf
}
