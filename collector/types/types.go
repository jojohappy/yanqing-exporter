package types

type TcpExtStat struct {
	PruneCalled        uint64
	LockDroppedIcmps   uint64
	ArpFilter          uint64
	TW                 uint64
	DelayedACKLocked   uint64
	ListenOverflows    uint64
	ListenDrops        uint64
	TCPPrequeueDropped uint64
	TCPTSReorder       uint64
	TCPDSACKUndo       uint64
	TCPLostRetransmit  uint64
	TCPLossFailures    uint64
	TCPFastRetrans     uint64
	TCPTimeouts        uint64
	TCPSchedulerFailed uint64
	TCPAbortOnMemory   uint64
	TCPAbortOnTimeout  uint64
	TCPAbortFailed     uint64
	TCPMemoryPressures uint64
	TCPSpuriousRTOs    uint64
	TCPBacklogDrop     uint64
	TCPMinTTLDrop      uint64
}
