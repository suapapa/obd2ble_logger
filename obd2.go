package main

func getSupportedPids(val []byte) []byte {
	var pids []byte

	currPid := byte(0x01)
	for i := 0; i < len(val); i++ {
		for b := 7; b >= 0; b-- {
			if val[i]&(1<<b) != 0 {
				pids = append(pids, currPid)
			}
			currPid++
		}
	}

	return pids
}
