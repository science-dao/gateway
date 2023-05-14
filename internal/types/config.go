package types

type Config struct {
	NodeURL     string `mapstructure:"NODE_URL"`
	DaoAddress  string `mapstructure:"DAO_ADDRESS"`
	Mode        string `mapstructure:"MODE"`
	ChainId     int64  `mapstructure:"CHAIN_ID"`
	Key         string `mapsturcture:"KEY"`
	Port        int    `mapstructure:"PORT"`
	LotusPath   string `mapstructure:"LOTUS_PATH"`
	AccessToken string `mapstructure:"ACCESS_TOKEN_ADDRESS"`
	W3SKey      string `mapstructure:"W3S_KEY"`
}
