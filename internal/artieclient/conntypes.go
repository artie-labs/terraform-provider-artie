package artieclient

import "fmt"

type ConnectorType string

const (
	BigQuery   ConnectorType = "bigquery"
	DynamoDB   ConnectorType = "dynamodb"
	MongoDB    ConnectorType = "mongodb"
	MySQL      ConnectorType = "mysql"
	MSSQL      ConnectorType = "mssql"
	Oracle     ConnectorType = "oracle"
	PostgreSQL ConnectorType = "postgresql"
	Redshift   ConnectorType = "redshift"
	S3         ConnectorType = "s3"
	Snowflake  ConnectorType = "snowflake"
)

var AllSourceTypes = []string{
	string(DynamoDB),
	string(MongoDB),
	string(MySQL),
	string(MSSQL),
	string(Oracle),
	string(PostgreSQL),
}

var AllDestinationTypes = []string{
	string(BigQuery),
	string(MSSQL),
	string(Redshift),
	string(S3),
	string(Snowflake),
}

func ConnectorTypeFromString(connType string) (ConnectorType, error) {
	switch ConnectorType(connType) {
	case BigQuery:
		return BigQuery, nil
	case DynamoDB:
		return DynamoDB, nil
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
	default:
		return "", fmt.Errorf("invalid connector type: %s", connType)
	}
}
