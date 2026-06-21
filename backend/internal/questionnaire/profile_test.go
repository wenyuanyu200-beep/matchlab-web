package questionnaire

import (
	"reflect"
	"strings"
	"testing"
)

func TestGenerateProfile(t *testing.T) {
	answers := Answers{
		Interests:          []string{"电赛", "STM32", "硬件", "电赛"},
		Skills:             []string{"嵌入式", "焊接"},
		AvailableTime:      "周末下午",
		ActivityTypes:      []string{"competition", "project"},
		Goal:               "找队友一起参加比赛",
		CommunicationStyle: "稳定沟通",
	}

	profile := GenerateProfile("activity", answers)
	wantTags := StringList{"电赛", "STM32", "硬件", "嵌入式", "焊接", "competition", "project"}
	if !reflect.DeepEqual(profile.Tags, wantTags) {
		t.Fatalf("tags=%v, want %v", profile.Tags, wantTags)
	}
	wantScores := ProfileScores{
		InterestScore:      80,
		SkillScore:         75,
		TimeScore:          70,
		GoalScore:          80,
		CommunicationScore: 75,
	}
	if profile.Scores != wantScores {
		t.Fatalf("scores=%#v, want %#v", profile.Scores, wantScores)
	}
	if !strings.Contains(profile.Summary, "电赛") || !strings.Contains(profile.Summary, "嵌入式") {
		t.Fatalf("summary does not describe signals: %s", profile.Summary)
	}
	for _, prohibited := range []string{"交友", "脱单", "约会"} {
		if strings.Contains(profile.Summary, prohibited) {
			t.Fatalf("summary contains prohibited term %q: %s", prohibited, profile.Summary)
		}
	}
}

func TestGenerateProfileNormalizesTagsAndUsesBaseScores(t *testing.T) {
	profile := GenerateProfile("activity", Answers{
		Interests:     []string{" STM32 ", "stm32", "", "  "},
		Skills:        []string{" 焊接 "},
		ActivityTypes: []string{"PROJECT", "project"},
	})

	wantTags := StringList{"STM32", "焊接", "PROJECT"}
	if !reflect.DeepEqual(profile.Tags, wantTags) {
		t.Fatalf("tags=%v, want %v", profile.Tags, wantTags)
	}
	if profile.Scores.TimeScore != 20 || profile.Scores.GoalScore != 20 || profile.Scores.CommunicationScore != 20 {
		t.Fatalf("missing answers should receive base scores: %#v", profile.Scores)
	}
}

func TestGenerateProfileDoesNotEchoProhibitedTerms(t *testing.T) {
	profile := GenerateProfile("activity", Answers{
		Interests: StringList{"交友活动", "硬件"},
		Skills:    StringList{"约会策划", "焊接"},
	})
	for _, prohibited := range []string{"交友", "脱单", "约会"} {
		if strings.Contains(profile.Summary, prohibited) {
			t.Fatalf("summary contains prohibited term %q: %s", prohibited, profile.Summary)
		}
	}
}
