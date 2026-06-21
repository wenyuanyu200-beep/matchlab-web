package match

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/questionnaire"
)

func TestScoreActivityUsesRequestedWeights(t *testing.T) {
	signals := UserSignals{Answers: questionnaire.Answers{
		Interests:     questionnaire.StringList{"电赛", "STM32", "硬件"},
		Skills:        questionnaire.StringList{"嵌入式", "焊接"},
		ActivityTypes: questionnaire.StringList{"competition"},
		AvailableTime: "周末下午",
		Goal:          "电赛",
	}}
	act := activity.Activity{
		ID:            uuid.New(),
		Type:          "competition",
		Title:         "电赛组队",
		Description:   "参加比赛",
		Tags:          activity.StringList{"电赛", "STM32", "硬件"},
		PreferredTags: activity.StringList{"嵌入式", "焊接"},
		TimeText:      "周末下午",
	}

	got := ScoreActivity(signals, act)
	want := DetailScores{Interest: 30, Skill: 25, Type: 20, Time: 10, Goal: 15}
	if got.DetailScores != want || got.Score != 100 {
		t.Fatalf("score=%d details=%#v, want 100 %#v", got.Score, got.DetailScores, want)
	}
	if !strings.Contains(got.Reason, "电赛") || !strings.Contains(got.Reason, "嵌入式") {
		t.Fatalf("reason does not explain matches: %s", got.Reason)
	}
}

func TestScoreActivityHandlesPartialAndCaseInsensitiveMatches(t *testing.T) {
	signals := UserSignals{Answers: questionnaire.Answers{
		Interests:     questionnaire.StringList{"STM32", "硬件", "控制"},
		Skills:        questionnaire.StringList{"Embedded", "焊接"},
		ActivityTypes: questionnaire.StringList{"PROJECT"},
		AvailableTime: "周末 下午",
		Goal:          "智能硬件项目",
	}}
	act := activity.Activity{
		Type:          "project",
		Title:         "智能硬件协作",
		Description:   "共同完成项目原型",
		Tags:          activity.StringList{"stm32", "硬件"},
		PreferredTags: activity.StringList{"embedded"},
		TimeText:      "周末上午",
	}

	got := ScoreActivity(signals, act)
	if got.DetailScores.Interest != 20 || got.DetailScores.Skill != 13 || got.DetailScores.Type != 20 {
		t.Fatalf("unexpected overlap scores: %#v", got.DetailScores)
	}
	if got.DetailScores.Time <= 0 || got.DetailScores.Time >= 10 || got.DetailScores.Goal <= 0 {
		t.Fatalf("expected partial text scores: %#v", got.DetailScores)
	}
}

func TestScoreActivityNoOverlapUsesNeutralSafeReason(t *testing.T) {
	got := ScoreActivity(UserSignals{}, activity.Activity{Type: "lecture", Title: "学术讲座"})
	if got.Score != 0 {
		t.Fatalf("score=%d, want 0", got.Score)
	}
	if !strings.HasPrefix(got.Reason, "推荐原因：") {
		t.Fatalf("unexpected reason: %s", got.Reason)
	}
	for _, prohibited := range []string{"交友", "脱单", "约会"} {
		if strings.Contains(got.Reason, prohibited) {
			t.Fatalf("reason contains prohibited term %q: %s", prohibited, got.Reason)
		}
	}
}

func TestRankActivitiesSortsByScoreThenCreationTimeAndLimits(t *testing.T) {
	now := time.Now().UTC()
	signals := UserSignals{Answers: questionnaire.Answers{Interests: questionnaire.StringList{"硬件"}}}
	activities := []activity.Activity{
		{ID: uuid.New(), Title: "低分", Tags: activity.StringList{"软件"}, CreatedAt: now.Add(time.Hour)},
		{ID: uuid.New(), Title: "较早高分", Tags: activity.StringList{"硬件"}, CreatedAt: now},
		{ID: uuid.New(), Title: "较新高分", Tags: activity.StringList{"硬件"}, CreatedAt: now.Add(time.Minute)},
	}

	got := RankActivities(signals, activities, 2)
	if len(got) != 2 || got[0].Activity.Title != "较新高分" || got[1].Activity.Title != "较早高分" {
		t.Fatalf("unexpected ranking: %#v", got)
	}
}
