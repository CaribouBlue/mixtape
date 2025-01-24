package session

import (
	"sort"
	"time"
)

type SessionPhase string

const (
	SubmissionPhase SessionPhase = "submission"
	VotePhase       SessionPhase = "vote"
	ResultPhase     SessionPhase = "result"
)

type Submission struct {
	Id      string `json:"id"`
	UserId  int64  `json:"userId"`
	TrackId string `json:"trackId"`
}

type Vote struct {
	Id           string `json:"id"`
	UserId       int64  `json:"userId"`
	SubmissionId string `json:"submissionId"`
}

type Playlist struct {
	Id     string `json:"id"`
	UserId int64  `json:"userId"`
}

type Result struct {
	SubmissionId string
	VoteCount    int
	Place        int
}

type ByVoteCountDesc []Result

func (a ByVoteCountDesc) Len() int           { return len(a) }
func (a ByVoteCountDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVoteCountDesc) Less(i, j int) bool { return a[i].VoteCount > a[j].VoteCount }

type Session struct {
	Id                 int64         `json:"id"`
	Name               string        `json:"name"`
	Submissions        []Submission  `json:"submissions"`
	Votes              []Vote        `json:"votes"`
	Playlists          []Playlist    `json:"playlists"`
	Results            []Result      `json:"result"`
	CreatedAt          time.Time     `json:"createdAt"`
	MaxSubmissions     int           `json:"maxSubmissions"`
	StartAt            time.Time     `json:"startAt"`
	SubmissionDuration time.Duration `json:"submissionDuration"`
	VoteDuration       time.Duration `json:"voteDuration"`
}

func NewSession(name string) *Session {
	dayDuration := time.Duration(24) * time.Hour

	return &Session{
		Name:               name,
		Submissions:        []Submission{},
		Votes:              []Vote{},
		Playlists:          []Playlist{},
		CreatedAt:          time.Now(),
		MaxSubmissions:     5,
		StartAt:            time.Now(),
		SubmissionDuration: time.Duration(10 * dayDuration),
		VoteDuration:       time.Duration(10 * dayDuration),
	}
}

func (s *Session) Phase() SessionPhase {
	if time.Since(s.StartAt) < s.SubmissionDuration {
		return SubmissionPhase
	}

	if time.Since(s.StartAt) < s.SubmissionDuration+s.VoteDuration {
		return VotePhase
	}

	return ResultPhase
}

func (s *Session) RemainingPhaseDuration() time.Duration {
	sessionPhase := s.Phase()

	switch {
	case sessionPhase == SubmissionPhase:
		return s.SubmissionDuration - time.Since(s.StartAt)
	case sessionPhase == VotePhase:
		return s.SubmissionDuration + s.VoteDuration - time.Since(s.StartAt)
	default:
		return 0
	}
}

func (s *Session) GetResults() []Result {
	if len(s.Results) > 0 {
		return s.Results
	}

	voteCountBySubmission := make(map[string]int)
	for _, vote := range s.Votes {
		voteCountBySubmission[vote.SubmissionId]++
	}

	s.Results = make([]Result, 0)
	for submissionId, count := range voteCountBySubmission {
		s.Results = append(s.Results, Result{
			SubmissionId: submissionId,
			VoteCount:    count,
		})
	}

	sort.Sort(ByVoteCountDesc(s.Results))

	place := 1
	currentBest := s.Results[0].VoteCount
	for i, result := range s.Results {
		if result.VoteCount < currentBest {
			place += 1
			currentBest = result.VoteCount
		}
		s.Results[i].Place = place
	}

	return s.Results
}
