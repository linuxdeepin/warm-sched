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

func CaptureByPIDs(pids []int, handle FileInfoHandleFunc) error {
	fs, err := ReferencedFilesByPID(pids...)
	if err != nil {
		return err
	}
	for _, fname := range fs {
		finfo, err := FileMincore(fname)
		if err != nil {
			continue
		}
		if err := handle(finfo); err != nil {
			return err
		}
	}
	return nil
}

func CaptureByFileList(list []string, _ bool, handle FileInfoHandleFunc) error {
	for _, fname := range list {
		finfo, err := FileMincore(fname)
		if err != nil {
			continue
		}
		if err := handle(finfo); err != nil {
			return err
		}
	}
	return nil
}
