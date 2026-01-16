package artieclient

import "fmt"

type ConnectorType string

const (
	API         ConnectorType = "api"
	BigQuery    ConnectorType = "bigquery"
	CockroachDB ConnectorType = "cockroach"
	DynamoDB    ConnectorType = "dynamodb"
	GCS         ConnectorType = "gcs"
	Iceberg     ConnectorType = "iceberg"
	MongoDB     ConnectorType = "mongodb"
	MySQL       ConnectorType = "mysql"
	MSSQL       ConnectorType = "mssql"
	Oracle      ConnectorType = "oracle"
	PostgreSQL  ConnectorType = "postgresql"
	Redshift    ConnectorType = "redshift"
	S3          ConnectorType = "s3"
	Snowflake   ConnectorType = "snowflake"
	Databricks  ConnectorType = "databricks"
)

var AllSourceTypes = []string{
	string(API),
	string(CockroachDB),
	string(DynamoDB),
	string(MongoDB),
	string(MySQL),
	string(MSSQL),
	string(Oracle),
	string(PostgreSQL),
}

var AllDestinationTypes = []string{
	string(BigQuery),
	string(GCS),
	string(Iceberg),
	string(MSSQL),
	string(Redshift),
	string(S3),
	string(Snowflake),
	string(Databricks),
}

var AllConnectorTypes = append(AllSourceTypes, AllDestinationTypes...)

func ConnectorTypeFromString(connType string) (ConnectorType, error) {
	switch ConnectorType(connType) {
	case API:
		return API, nil
	case BigQuery:
		return BigQuery, nil
	case CockroachDB:
		return CockroachDB, nil
	case DynamoDB:
		return DynamoDB, nil
	case GCS:
		return GCS, nil
	case Iceberg:
		return Iceberg, nil
	case MongoDB:
		return MongoDB, nil
	case MySQL:
		return MySQL, nil
	case MSSQL:
		return MSSQL, nil
	case Oracle:
		return Oracle, nil
	case PostgreSQL:
		return PostgreSQL, nil
	case Redshift:
		return Redshift, nil
	case S3:
		return S3, nil
	case Snowflake:
		return Snowflake, nil
	case Databricks:
		return Databricks, nil
	default:
		return "", fmt.Errorf("invalid connector type: %s", connType)
	}
}
