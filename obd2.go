package main

const (
	OBD2Svc01EngineSpeed  = 0x0C
	OBD2Svc01VehicleSpeed = 0x0D
)

func parseVal2SupportedPids(val []byte) []byte {
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
