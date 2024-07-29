package models

type DestinationAPIModel struct {
	UUID          string                          `json:"uuid"`
	CompanyUUID   string                          `json:"companyUUID"`
	Name          string                          `json:"name"`
	Label         string                          `json:"label"`
	LastUpdatedAt string                          `json:"lastUpdatedAt"`
	SSHTunnelUUID string                          `json:"sshTunnelUUID"`
	Config        DestinationSharedConfigAPIModel `json:"sharedConfig"`
}

type DestinationSharedConfigAPIModel struct {
	Host                string `json:"host"`
	Port                int64  `json:"port"`
	Endpoint            string `json:"endpoint"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	GCPProjectID        string `json:"projectID"`
	GCPLocation         string `json:"location"`
	GCPCredentialsData  string `json:"credentialsData"`
	AWSAccessKeyID      string `json:"awsAccessKeyID"`
	AWSSecretAccessKey  string `json:"awsSecretAccessKey"`
	AWSRegion           string `json:"awsRegion"`
	SnowflakeAccountURL string `json:"accountURL"`
	SnowflakeVirtualDWH string `json:"virtualDWH"`
	SnowflakePrivateKey string `json:"privateKey"`
}
