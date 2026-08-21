package broker

type KafkaConfig struct {
	Username        string     `json:"username" mapstructure:"username"`
	Password        string     `json:"password" mapstructure:"password"`
	Brokers         string     `json:"brokers" mapstructure:"brokers"`
	CommitType      CommitType `json:"commit_type" mapstructure:"commit_type"`
	AutoCreateTopic bool       `json:"auto_create_topic" mapstructure:"auto_create_topic"`
}
