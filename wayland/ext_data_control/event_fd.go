package ext_data_control

func (i *ExtDataControlSourceV1) EventTakesFd(opcode uint32) bool {
	return opcode == 0 // ext_data_control_source_v1.send
}
