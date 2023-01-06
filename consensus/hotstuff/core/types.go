package core

type MsgType uint64

const (
	MsgTypeNewView       MsgType = 1
	MsgTypePrepare       MsgType = 2
	MsgTypePrepareVote   MsgType = 3
	MsgTypePreCommit     MsgType = 4
	MsgTypePreCommitVote MsgType = 5
	MsgTypeCommit        MsgType = 6
	MsgTypeCommitVote    MsgType = 7
	MsgTypeDecide        MsgType = 8
)

func (m MsgType) String() string {
	switch m {
	case MsgTypeNewView:
		return "NEW_VIEW"
	case MsgTypePrepare:
		return "PREPARE"
	case MsgTypePrepareVote:
		return "PREPARE_VOTE"
	case MsgTypePreCommit:
		return "PRECOMMIT"
	case MsgTypePreCommitVote:
		return "PRECOMMIT_VOTE"
	case MsgTypeCommit:
		return "COMMIT"
	case MsgTypeCommitVote:
		return "COMMIT_VOTE"
	case MsgTypeDecide:
		return "DECIDE"
	default:
		return "UNKNOWN"
	}
}

func (m MsgType) Value() uint64 {
	return uint64(m)
}
