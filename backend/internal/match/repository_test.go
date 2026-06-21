package match

import (
	"database/sql/driver"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

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

func TestMatchUpsertOnlyReturnsID(t *testing.T) {
	clauses := matchUpsertClauses()
	if len(clauses) != 2 {
		t.Fatalf("clauses=%#v", clauses)
	}
	returning, ok := clauses[1].(clause.Returning)
	if !ok {
		t.Fatalf("second clause=%T want clause.Returning", clauses[1])
	}
	if len(returning.Columns) != 1 || returning.Columns[0].Name != "id" {
		t.Fatalf("returning columns=%#v want id only", returning.Columns)
	}
}

func TestDetailScoresJSONBRoundTrip(t *testing.T) {
	want := DetailScores{Interest: 26, Skill: 22, Type: 20, Time: 8, Goal: 12}
	value, err := want.Value()
	if err != nil {
		t.Fatalf("value: %v", err)
	}
	encoded, ok := value.(string)
	if !ok {
		t.Fatalf("value type=%T want string", value)
	}
	if strings.Contains(encoded, "::jsonb") {
		t.Fatalf("value=%q must not contain PostgreSQL cast", encoded)
	}
	if encoded != `{"interest":26,"skill":22,"type":20,"time":8,"goal":12}` {
		t.Fatalf("value=%q", encoded)
	}
	var got DetailScores
	if err := got.Scan(value); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%#v want=%#v", got, want)
	}
}

func TestDetailScoresValueReturnsPlainEmptyJSON(t *testing.T) {
	value, err := (DetailScores{}).Value()
	if err != nil {
		t.Fatalf("value: %v", err)
	}
	encoded, ok := value.(string)
	if !ok {
		t.Fatalf("value type=%T want string", value)
	}
	if encoded != "{}" {
		t.Fatalf("value=%q want {}", encoded)
	}
	if strings.Contains(encoded, "::jsonb") {
		t.Fatalf("value=%q must not contain PostgreSQL cast", encoded)
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
		{name: "empty json bytes", value: []byte(`{}`), want: DetailScores{}},
		{name: "empty json string", value: `{}`, want: DetailScores{}},
		{name: "json bytes", value: []byte(`{"interest":26,"skill":22,"type":20,"time":8,"goal":12}`), want: want},
		{name: "json string", value: `{"interest":26,"skill":22,"type":20,"time":8,"goal":12}`, want: want},
		{name: "double encoded json", value: `"{\"interest\":26,\"skill\":22,\"type\":20,\"time\":8,\"goal\":12}"`, want: want},
		{name: "legacy quoted jsonb cast", value: `'{}'::jsonb`, want: DetailScores{}},
		{name: "legacy quoted jsonb cast bytes", value: []byte(`'{}'::jsonb`), want: DetailScores{}},
		{name: "legacy jsonb cast", value: `{}::jsonb`, want: DetailScores{}},
		{name: "legacy trailing quote", value: `{}'`, want: DetailScores{}},
		{name: "legacy quoted json", value: `'{}'`, want: DetailScores{}},
		{name: "legacy quoted full jsonb cast", value: `'{"interest":30,"skill":25,"type":20,"time":8,"goal":12}'::jsonb`, want: DetailScores{Interest: 30, Skill: 25, Type: 20, Time: 8, Goal: 12}},
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
