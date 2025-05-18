package irsdk

type Variable struct {
	VarType     VarType // irsdk_VarType
	Offset      int     // offset fron start of buffer row
	Count       int     // number of entrys (array) so length in bytes would be irsdk_VarTypeBytes[type] * count
	CountAsTime bool
	Name        string
	Desc        string
	Unit        string
	Values      []any
}
