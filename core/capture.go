package core

func CaptureByMincores(mps []string, handle FileInfoHandleFunc) error {
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

func CaptureByPIDs(pids []int, acc interface{}, handle FileInfoHandleFunc) error {
	panic("Not Implement")
}
