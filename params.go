package rdb

// If the N (Name) field is not specified is not specified, then the order
// of the parameter should be used if the driver supports it.
type Param struct {
	N string  // Optional Parameter Name.
	T SqlType // Parameter Type.
	L int     // Paremeter Length.

	// Value for input parameter.
	// If the value is an io.Reader it will read the value directly to the wire.
	// If this satisfies the Filler interface the value will be fetched from
	// that interface.
	V interface{}

	Null      bool
	Scale     int
	Precision int
}

// If the input parameter value isn't populated in the command,
// the value can be filled in at the time of query.
// If the N (Name) field is not specified, then the order of the
// parameters or values are used if the driver supports it.
type Value struct {
	N string // Parameter Name.

	// Value for input parameter.
	// If the value is an io.Reader it will read the value directly to the wire.
	V interface{}

	Param *Param
}

type SqlColumn struct {
	Name     string
	Index    int
	SqlType  uint8
	Length   uint32
	Unlimit  bool // Provides near unlimited length.
	Nullable bool
}

// If the command output fields are specified, the Field output can help manage
// how the result rows are copied to.
type Field struct {
	N         string     // Optional Field Name.
	Type      NativeType // TODO: Should this stay or go?
	NullValue interface{}
}

type IsolationLevel byte

const (
	IsoLevelDefault IsolationLevel = iota
	IsoLevelReadUncommited
	IsoLevelReadCommited
	IsoLevelWriteCommited
	IsoLevelRepeatableRead
	IsoLevelSerializable
	IsoLevelSnapshot
)

type Arity byte

// The number of rows to expect from a command.
const (
	Many Arity = iota
	One
	Zero
	OneOnly
	ZeroOnly
)

// Command represents a SQL command and can be used from many different
// queries at the same time, so long as the input parameter values
// "Input[N].V (Value)" are not set in the Param struct but passed in with
// the actual query as Value.
type Command struct {
	// The SQL to be used in the command.
	Sql string

	// Number of rows expected.
	//   If Arity is One or OneOnly, only the first row is returned.
	//   If Arity is OneOnly, if more results are returned an error is returned.
	//   If Arity is Zero or ZeroOnly, no rows are returned.
	//   If Arity is ZeroOnnly, if any results are returned an error is returned.
	Arity Arity
	Input []*Param

	// Optional fields to specify output marshal.
	Output []*Field

	// If set to true silently truncates text longer then the field.
	// If this is set to false text truncation will result in an error.
	TruncLongText bool

	// Optional name of the command. May be used if logging.
	Name string
}

// The table schema and properties.
type Schema struct {
	Columns []*SqlColumn
}
