package model

type SystemConfig struct {
	Key   string `gorm:"type:text;primaryKey" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
