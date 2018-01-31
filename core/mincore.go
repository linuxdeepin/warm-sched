package core

func TakeByMincores(mps []string, handle FileInfoTakeHandleFunc) error {
	if err := supportProduceByKernel(); err != nil {
		return err
	}
	ch := make(chan FileInfo)

	go generateFileInfoByKernel(ch, mps)

	for info := range ch {
		if err := handle(info); err != nil {
			return err
		}
	}
	return nil
}

func TakeByPIDs(pids []int, acc interface{}, handle FileInfoTakeHandleFunc) error {
	panic("Not Implement")
}
