package main

type Snapshot struct {
	infos []FileCacheInfo
}

func (s *Snapshot) Add(i FileCacheInfo) {
	s.infos = append(s.infos, i)
}

func TakeSnapshot(dirs []string) *Snapshot {
	ch := make(chan FileCacheInfo)
	snap := &Snapshot{}

	go Produce(ch, dirs)

	for info := range ch {
		snap.Add(info)
	}
	return snap
}

func (s *Snapshot) SaveTo(f string) error {
	panic("Not Implement SaveTo")
}
