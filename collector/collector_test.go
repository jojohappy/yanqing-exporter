package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

const TcpExtStatContent = `TcpExt: SyncookiesSent SyncookiesRecv SyncookiesFailed EmbryonicRsts PruneCalled RcvPruned OfoPruned OutOfWindowIcmps LockDroppedIcmps ArpFilter TW TWRecycled TWKilled PAWSPassive PAWSActive PAWSEstab DelayedACKs DelayedACKLocked DelayedACKLost ListenOverflows ListenDrops TCPPrequeued TCPDirectCopyFromBacklog TCPDirectCopyFromPrequeue TCPPrequeueDropped TCPHPHits TCPHPHitsToUser TCPPureAcks TCPHPAcks TCPRenoRecovery TCPSackRecovery TCPSACKReneging TCPFACKReorder TCPSACKReorder TCPRenoReorder TCPTSReorder TCPFullUndo TCPPartialUndo TCPDSACKUndo TCPLossUndo TCPLostRetransmit TCPRenoFailures TCPSackFailures TCPLossFailures TCPFastRetrans TCPForwardRetrans TCPSlowStartRetrans TCPTimeouts TCPLossProbes TCPLossProbeRecovery TCPRenoRecoveryFail TCPSackRecoveryFail TCPSchedulerFailed TCPRcvCollapsed TCPDSACKOldSent TCPDSACKOfoSent TCPDSACKRecv TCPDSACKOfoRecv TCPAbortOnData TCPAbortOnClose TCPAbortOnMemory TCPAbortOnTimeout TCPAbortOnLinger TCPAbortFailed TCPMemoryPressures TCPSACKDiscard TCPDSACKIgnoredOld TCPDSACKIgnoredNoUndo TCPSpuriousRTOs TCPMD5NotFound TCPMD5Unexpected TCPSackShifted TCPSackMerged TCPSackShiftFallback TCPBacklogDrop TCPMinTTLDrop TCPDeferAcceptDrop IPReversePathFilter TCPTimeWaitOverflow TCPReqQFullDoCookies TCPReqQFullDrop TCPRetransFail TCPRcvCoalesce TCPOFOQueue TCPOFODrop TCPOFOMerge TCPChallengeACK TCPSYNChallenge TCPFastOpenActive TCPFastOpenPassive TCPFastOpenPassiveFail TCPFastOpenListenOverflow TCPFastOpenCookieReqd TCPSpuriousRtxHostQueues BusyPollRxPackets
TcpExt: 0 0 1 0 4673 0 0 0 10001 10002 65134 0 0 0 0 0 412149 256 8109 10003 10004 70497 409689115 786173026 10005 67435855 844758 12718853 13857967 0 2561 0 42 1097 0 187 90 295 645 792 3404 0 43 29 23172 916 1951 692645 20273 12645 0 47 10007 23726 8087 69 15039 2 80502 834 10009 91 0 10050 0 0 88 9647 105 0 0 8320 14983 42486 5610 4563 0 8 0 0 0 0 36096311 486700 0 69 135 121 0 0 0 0 0 580258 0
IpExt: InNoRoutes InTruncatedPkts InMcastPkts OutMcastPkts InBcastPkts OutBcastPkts InOctets OutOctets InMcastOctets OutMcastOctets InBcastOctets OutBcastOctets InCsumErrors InNoECTPkts InECT1Pkts InECT0Pkts InCEPkts
IpExt: 0 0 172589 0 0 0 116957808614 76196277756 6213204 0 0 0 0 131510672 0 0 0`

const TcpStatContent = `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 157345035 1 ffff88e34d17b800 100 0 0 10 0
   1: FC1E16AC:ED36 F820000A:0050 01 00000000:00000000 02:0000003A 00000000  1000        0 157952987 2 ffff88e84ba91000 22 4 16 10 -1
   2: FC1E16AC:8390 F820000A:0050 01 00000000:00000000 02:00001CA3 00000000  1000        0 157963826 2 ffff88ec6efc9800 21 4 28 10 -1
   3: 0100007F:C62C 0100007F:1F90 06 00000000:00000000 03:0000142F 00000000     0        0 0 3 ffff88e880212ee0
   4: FC1E16AC:0050 2380000A:8F43 01 00000000:00000000 00:00000000 00000000 65534        0 159220417 1 ffff88e969041000 20 4 0 29 25
   5: FC1E16AC:0050 3180000A:28F9 01 00000000:00000000 00:00000000 00000000 65534        0 159220416 1 ffff88e969046000 20 4 0 33 31
   6: FC1E16AC:838C F820000A:0050 01 00000000:00000000 02:00000061 00000000  1000        0 157963824 2 ffff88ec6efcc800 22 4 28 10 -1
   7: FC1E16AC:90F0 F820000A:0050 01 00000000:00000000 02:00001D03 00000000  1000        0 157350646 2 ffff88eb724db000 21 4 24 7 7
   8: FC1E16AC:8394 F820000A:0050 01 00000000:00000000 02:00001D0A 00000000  1000        0 157963827 2 ffff88ec6efce000 20 4 28 10 -1
   9: FC1E16AC:ED38 F820000A:0050 01 00000000:00000000 02:00000000 00000000  1000        0 157952988 2 ffff88e84ba97000 21 4 22 10 -1
  10: FC1E16AC:839E F820000A:0050 01 00000000:00000000 02:0000003F 00000000  1000        0 157963829 2 ffff88ec6efca800 22 4 30 10 -1
  11: FC1E16AC:C4D4 F820000A:0050 01 00000000:00000000 02:0000003C 00000000  1000        0 157381529 2 ffff88ece3621000 21 4 20 10 -1
  12: FC1E16AC:E5E6 F820000A:0050 01 00000000:00000000 02:00001CD0 00000000  1000        0 157350175 2 ffff88eb724d8000 21 4 22 7 7
  13: FC1E16AC:E7E6 F820000A:0050 01 00000000:00000000 02:00000003 00000000  1000        0 157348221 2 ffff88e598cf7000 21 4 30 10 -1
  14: FC1E16AC:0050 2380000A:8523 06 00000000:00000000 03:00001564 00000000     0        0 0 3 ffff88e69eba7760
  15: FC1E16AC:838E F820000A:0050 01 00000000:00000000 02:00001CD3 00000000  1000        0 157963825 2 ffff88ec6efce800 21 4 28 10 -1
  16: FC1E16AC:EB3A F820000A:0050 01 00000000:00000000 02:00001D00 00000000  1000        0 157343447 2 ffff88e468d89800 22 4 22 10 -1
  17: FC1E16AC:83A0 F820000A:0050 01 00000000:00000000 02:00000041 00000000  1000        0 157963830 2 ffff88eb9a692000 22 4 30 10 -1
  18: FC1E16AC:8396 F820000A:0050 01 00000000:00000330 02:00001CA3 00000000  1000        0 157963828 3 ffff88ec6efc9000 21 4 28 10 -1
  19: FC1E16AC:ED32 F820000A:0050 01 00000000:00000000 02:00001CD0 00000000  1000        0 157952986 2 ffff88e84ba96800 22 4 14 10 -1
  20: FC1E16AC:0050 3180000A:1DD3 06 00000000:00000000 03:00001562 00000000     0        0 0 3 ffff88e841beccc0
  21: FC1E16AC:E5E8 F820000A:0050 01 00000000:00000000 02:00000000 00000000  1000        0 157350176 2 ffff88eb724dc800 22 4 22 10 -1
  22: FC1E16AC:0050 1C80000A:ECFB 01 00000000:00000000 00:00000000 00000000 65534        0 159202104 2 ffff88e9d0296000 20 4 0 54 34`

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

func TestTcpStatCollect(t *testing.T) {
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
	tcpStatFile := path.Join(yqStatDir, "/proc/1/net/tcp")

	if err = ioutil.WriteFile(tcpStatFile, []byte(TcpStatContent), 0644); err != nil {
		t.Fatal(err)
	}

	tcpStat, tcpStatPort, err := tcpStatsFromProc(yqStatDir, 1, "net/tcp")
	fmt.Println(tcpStat)
	fmt.Println(tcpStatPort)

	for port, stats := range tcpStatPort.Stats {
		fmt.Println(strconv.FormatInt(port, 10))
		fmt.Println(stats.Established)
	}
}
