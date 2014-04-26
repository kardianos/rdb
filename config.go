package rdb

type Config struct {
	DriverName string

	Username string
	Password string
	Url      string
	Port     int
	Instance string

	KV map[string]interface{}
}

// Provides a standard method to parse configuration options from a text.
// The instance field can also hold the filename in case of a file based connection.
//   driver://[username:password@][url[:port]]/[Instance]?opt1=valA&opt2=valB
//   sqlite:///C:/folder/file.sqlite3?opt1=valA&opt2=valB
//   sqlite:///srv/folder/file.sqlite3?opt1=valA&opt2=valB
func ParseConfig(url string) (*Config, error) {
	return nil, nil
}
func ParseConfigM(url string) *Config {
	return nil
}

type DriverOption struct {
	Name string

	Description  string
	Type         Type
	DefaultValue interface{}
}

type DriverSupport struct {
	NamedParameter   bool // Supports named parameters.
	FluidType        bool // Like SQLite.
	MultipleResult   bool // Supports returning multiple result sets.
	SecureConnection bool // Supports a secure connection.
	BulkInsert       bool // Supports a fast bulk insert method.
	Notification     bool // Supports driver notifications.
	UserDataTypes    bool // Handles user supplied data types.
}

type DriverMeta struct {
	Options []*DriverOption
	DriverSupport
}

func DriverMetaInfo(driverName string) *DriverMeta {
	return nil
}
