package match

import (
	"database/sql/driver"
	"reflect"
	"strings"
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

func TestDetailScoresScanHandlesDriverShapes(t *testing.T) {
	want := DetailScores{Interest: 26, Skill: 22, Type: 20, Time: 8, Goal: 12}
	tests := []struct {
		name  string
		value any
		want  DetailScores
	}{
		{name: "nil", value: nil, want: DetailScores{}},
		{name: "empty bytes", value: []byte{}, want: DetailScores{}},
		{name: "json bytes", value: []byte(`{"interest":26,"skill":22,"type":20,"time":8,"goal":12}`), want: want},
		{name: "json string", value: `{"interest":26,"skill":22,"type":20,"time":8,"goal":12}`, want: want},
		{name: "double encoded json", value: `"{\"interest\":26,\"skill\":22,\"type\":20,\"time\":8,\"goal\":12}"`, want: want},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got DetailScores
			if err := got.Scan(test.value); err != nil {
				t.Fatalf("scan: %v", err)
			}
			if got != test.want {
				t.Fatalf("got=%#v want=%#v", got, test.want)
			}
		})
	}
}

func TestDetailScoresScanRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		message string
	}{
		{name: "unsupported type", value: driver.Value(123), message: "unsupported DetailScores scan type"},
		{name: "invalid json", value: `{bad json`, message: "scan DetailScores: invalid json"},
		{name: "invalid double encoded json", value: `"not json"`, message: "scan DetailScores: invalid json"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got DetailScores
			err := got.Scan(test.value)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("err=%v want message %q", err, test.message)
			}
		})
	}
}
