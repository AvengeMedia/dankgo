package client

func (i *DataSource) EventTakesFd(opcode uint32) bool {
	return opcode == 1
}

func (i *Keyboard) EventTakesFd(opcode uint32) bool {
	return opcode == 0
}
