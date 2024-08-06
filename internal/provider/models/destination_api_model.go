package models

type DestinationAPIModel struct {
	UUID          string                          `json:"uuid"`
	Type          string                          `json:"name"`
	Label         string                          `json:"label"`
	SSHTunnelUUID string                          `json:"sshTunnelUUID"`
	Config        DestinationSharedConfigAPIModel `json:"sharedConfig"`
}

type DestinationSharedConfigAPIModel struct {
	Host                string `json:"host"`
	Port                int32  `json:"port"`
	Endpoint            string `json:"endpoint"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	GCPProjectID        string `json:"projectID"`
	GCPLocation         string `json:"location"`
	GCPCredentialsData  string `json:"credentialsData"`
	SnowflakeAccountURL string `json:"accountURL"`
	SnowflakeVirtualDWH string `json:"virtualDWH"`
	SnowflakePrivateKey string `json:"privateKey"`
}
