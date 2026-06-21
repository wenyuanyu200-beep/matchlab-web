package match

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"matchlab/backend/internal/activity"
)

func TestNewRecordUsesDirectRecommendationFields(t *testing.T) {
	userID := uuid.New()
	questionnaireID := uuid.New()
	activityID := uuid.New()
	now := time.Now().UTC()
	recommendation := Recommendation{
		Activity:     activity.Activity{ID: activityID},
		Score:        88,
		DetailScores: DetailScores{Interest: 26, Skill: 22, Type: 20, Time: 8, Goal: 12},
		Reason:       "推荐原因：活动方向与画像较为匹配。",
	}

	record := newRecord(userID, questionnaireID, recommendation, now)

	if record.TargetID != activityID || record.TargetType != "activity" || record.Algorithm != "rules" || record.AlgorithmVersion != AlgorithmVersion {
		t.Fatalf("unexpected target metadata: %#v", record)
	}
	if record.ActivityID != activityID || record.Score != 88 || record.DetailScores != recommendation.DetailScores || record.Reason != recommendation.Reason {
		t.Fatalf("unexpected recommendation fields: %#v", record)
	}
	if record.QuestionnaireID == nil || *record.QuestionnaireID != questionnaireID || !record.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected references/timestamps: %#v", record)
	}
}

func TestDetailScoresJSONBRoundTrip(t *testing.T) {
	want := DetailScores{Interest: 26, Skill: 22, Type: 20, Time: 8, Goal: 12}
	value, err := want.Value()
	if err != nil {
		t.Fatalf("value: %v", err)
	}
	var got DetailScores
	if err := got.Scan(value); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%#v want=%#v", got, want)
	}
}
