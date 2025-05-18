package irsdk

type VarType uint8

const (
	VarTypeChar     VarType = 0
	VarTypeBool     VarType = 1
	VarTypeInt      VarType = 2
	VarTypeBitField VarType = 3
	VarTypeFloat    VarType = 4
	VarTypeDouble   VarType = 5
	VarTypeETCount  VarType = 6
)
