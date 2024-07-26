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
	GCPProjectID        string `json:"projectID"`
	GCPLocation         string `json:"location"`
	AWSAccessKeyID      string `json:"awsAccessKeyID"`
	AWSRegion           string `json:"awsRegion"`
	SnowflakeAccountURL string `json:"accountURL"`
	SnowflakeVirtualDWH string `json:"virtualDWH"`
	// TODO sensitive fields
	// Password           string `json:"password"`
	// GCPCredentialsData string `json:"credentialsData"`
	// AWSSecretAccessKey string `json:"awsSecretAccessKey"`
	// SnowflakePrivateKey string `json:"privateKey"`
}
