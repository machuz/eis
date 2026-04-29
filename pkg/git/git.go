package git

import "github.com/machuz/eis/v2/internal/git"

type Commit = git.Commit
type FileStat = git.FileStat
type BlameLine = git.BlameLine

var (
	ParseLog                     = git.ParseLog
	ParseMergeCommits            = git.ParseMergeCommits
	ListFiles                    = git.ListFiles
	ListAllFiles                 = git.ListAllFiles
	ListFilesAtCommit            = git.ListFilesAtCommit
	BlameFile                    = git.BlameFile
	BlameFileAtCommit            = git.BlameFileAtCommit
	BlameFileAtParent            = git.BlameFileAtParent
	ConcurrentBlameFiles         = git.ConcurrentBlameFiles
	ConcurrentBlameFilesAtCommit = git.ConcurrentBlameFilesAtCommit
	DiffTreeFiles                = git.DiffTreeFiles
	FilterFilesBySize            = git.FilterFilesBySize
	FindCommitAtDate             = git.FindCommitAtDate
	HeadHash                     = git.HeadHash
	IsShallowRepo                = git.IsShallowRepo
	SampleFiles                  = git.SampleFiles
)
