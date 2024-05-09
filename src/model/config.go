package model

type WhiteOrBlackList struct {
	Type      string   `json:"type" msgpack:"type"`
	WhiteList []string `json:"white_list" msgpack:"white_list"`
	BlackList []string `json:"black_list" msgpack:"black_list"`
}

type Game struct {
	ID                string           `json:"id" msgpack:"id"`
	Protocol          string           `json:"protocol" msgpack:"protocol"`
	Host              string           `json:"host" msgpack:"host"`
	Port              uint64           `json:"port" msgpack:"port"`
	ListenPort        uint64           `json:"listen_port" msgpack:"listen_port"`
	ExitNode          WhiteOrBlackList `json:"exit_node" msgpack:"exit_node"`
	EntryNode         WhiteOrBlackList `json:"entry_node" msgpack:"entry_node"`
	SpeedTestProtocol string           `json:"speed_test_protocol" msgpack:"speed_test_protocol"`
}

type System struct {
	ListenHost string `json:"listen_host" msgpack:"listen_host"`
}

type Config struct {
	Games  []Game `json:"games" msgpack:"games"`
	System System `json:"system" msgpack:"system"`
}

type SignedConfig struct {
	Config Config `json:"config" msgpack:"config"`
	Sign   string `json:"sign" msgpack:"sign"`
}

type FixedConfig struct {
	SuperAdminPubKey string   `json:"super_admin_pub_key" msgpack:"super_admin_pub_key"`
	ConfigPath       string   `json:"config_path" msgpack:"config_path"`
	DataPath         string   `json:"data_path" msgpack:"data_path"`
	BoostrapNodes    []string `json:"bootstrap_nodes" msgpack:"bootstrap_nodes"`
	Port             uint     `json:"port" msgpack:"port"`
}
