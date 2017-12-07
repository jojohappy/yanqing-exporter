package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const TcpExtStatContent = `TcpExt: SyncookiesSent SyncookiesRecv SyncookiesFailed EmbryonicRsts PruneCalled RcvPruned OfoPruned OutOfWindowIcmps LockDroppedIcmps ArpFilter TW TWRecycled TWKilled PAWSPassive PAWSActive PAWSEstab DelayedACKs DelayedACKLocked DelayedACKLost ListenOverflows ListenDrops TCPPrequeued TCPDirectCopyFromBacklog TCPDirectCopyFromPrequeue TCPPrequeueDropped TCPHPHits TCPHPHitsToUser TCPPureAcks TCPHPAcks TCPRenoRecovery TCPSackRecovery TCPSACKReneging TCPFACKReorder TCPSACKReorder TCPRenoReorder TCPTSReorder TCPFullUndo TCPPartialUndo TCPDSACKUndo TCPLossUndo TCPLostRetransmit TCPRenoFailures TCPSackFailures TCPLossFailures TCPFastRetrans TCPForwardRetrans TCPSlowStartRetrans TCPTimeouts TCPLossProbes TCPLossProbeRecovery TCPRenoRecoveryFail TCPSackRecoveryFail TCPSchedulerFailed TCPRcvCollapsed TCPDSACKOldSent TCPDSACKOfoSent TCPDSACKRecv TCPDSACKOfoRecv TCPAbortOnData TCPAbortOnClose TCPAbortOnMemory TCPAbortOnTimeout TCPAbortOnLinger TCPAbortFailed TCPMemoryPressures TCPSACKDiscard TCPDSACKIgnoredOld TCPDSACKIgnoredNoUndo TCPSpuriousRTOs TCPMD5NotFound TCPMD5Unexpected TCPSackShifted TCPSackMerged TCPSackShiftFallback TCPBacklogDrop TCPMinTTLDrop TCPDeferAcceptDrop IPReversePathFilter TCPTimeWaitOverflow TCPReqQFullDoCookies TCPReqQFullDrop TCPRetransFail TCPRcvCoalesce TCPOFOQueue TCPOFODrop TCPOFOMerge TCPChallengeACK TCPSYNChallenge TCPFastOpenActive TCPFastOpenPassive TCPFastOpenPassiveFail TCPFastOpenListenOverflow TCPFastOpenCookieReqd TCPSpuriousRtxHostQueues BusyPollRxPackets
TcpExt: 0 0 1 0 4673 0 0 0 10001 10002 65134 0 0 0 0 0 412149 256 8109 10003 10004 70497 409689115 786173026 10005 67435855 844758 12718853 13857967 0 2561 0 42 1097 0 187 90 295 645 792 3404 0 43 29 23172 916 1951 692645 20273 12645 0 47 10007 23726 8087 69 15039 2 80502 834 10009 91 0 10050 0 0 88 9647 105 0 0 8320 14983 42486 5610 4563 0 8 0 0 0 0 36096311 486700 0 69 135 121 0 0 0 0 0 580258 0
IpExt: InNoRoutes InTruncatedPkts InMcastPkts OutMcastPkts InBcastPkts OutBcastPkts InOctets OutOctets InMcastOctets OutMcastOctets InBcastOctets OutBcastOctets InCsumErrors InNoECTPkts InECT1Pkts InECT0Pkts InCEPkts
IpExt: 0 0 172589 0 0 0 116957808614 76196277756 6213204 0 0 0 0 131510672 0 0 0`

func TestTcpExtStatCollect(t *testing.T) {
	var err error
	baseDir := os.Getenv("TEST_YQ_DIR")
	if len(baseDir) == 0 {
		baseDir = os.TempDir()
	}
	yqStatDir, err := ioutil.TempDir(baseDir, "yq_stat")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(yqStatDir); err != nil {
			t.Fatal(err)
		}
	}()
	os.MkdirAll(path.Join(yqStatDir, "/proc/1/net/"), 0755)
	tcpExtStatFile := path.Join(yqStatDir, "/proc/1/net/netstat")

	if err = ioutil.WriteFile(tcpExtStatFile, []byte(TcpExtStatContent), 0644); err != nil {
		t.Fatal(err)
	}

	tcpExtStat, err := scanTcpExtStats(yqStatDir, 1, "net/netstat")
	fmt.Println(tcpExtStat)
}
