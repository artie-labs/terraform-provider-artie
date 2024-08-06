package models

type DeploymentAPIResponse struct {
	Deployment DeploymentAPIModel `json:"deploy"`
}

type DeploymentAPIModel struct {
	UUID              string                    `json:"uuid"`
	Name              string                    `json:"name"`
	Status            string                    `json:"status"`
	DestinationUUID   string                    `json:"destinationUUID"`
	Source            SourceAPIModel            `json:"source"`
	DestinationConfig DestinationConfigAPIModel `json:"uniqueConfig"`
}

type SourceAPIModel struct {
	Type   string               `json:"name"`
	Config SourceConfigAPIModel `json:"config"`
	Tables []TableAPIModel      `json:"tables"`
}

type SourceConfigAPIModel struct {
	Host         string `json:"host"`
	SnapshotHost string `json:"snapshotHost"`
	Port         int32  `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Database     string `json:"database"`
}

type TableAPIModel struct {
	UUID                 string `json:"uuid"`
	Name                 string `json:"name"`
	Schema               string `json:"schema"`
	EnableHistoryMode    bool   `json:"enableHistoryMode"`
	IndividualDeployment bool   `json:"individualDeployment"`
	IsPartitioned        bool   `json:"isPartitioned"`
}

type DestinationConfigAPIModel struct {
	Dataset               string `json:"dataset"`
	Database              string `json:"database"`
	Schema                string `json:"schema"`
	SchemaOverride        string `json:"schemaOverride"`
	UseSameSchemaAsSource bool   `json:"useSameSchemaAsSource"`
	SchemaNamePrefix      string `json:"schemaNamePrefix"`
}
