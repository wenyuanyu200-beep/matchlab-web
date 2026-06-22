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
	wantTags := StringList{"电赛", "STM32", "硬件", "嵌入式", "焊接", "competition", "project", "稳定沟通", "找队友一起参加比赛"}
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

func TestGenerateProfileIncludesCollaborationDimensions(t *testing.T) {
	profile := GenerateProfile("activity", Answers{
		Interests:              StringList{"摄影", "公益"},
		Skills:                 StringList{"策划"},
		PreferredActivityTypes: StringList{"social", "volunteer"},
		TeamRole:               "组织协调者",
		WorkRhythm:             "提前规划",
		ParticipationMode:      "线下",
		CampusOrLocation:       "图书馆",
		MainGoal:               "参加校园活动",
		PartnerExpectation:     StringList{"认真负责"},
		ParticipationPurpose:   StringList{"丰富经历"},
	})

	for _, want := range []string{"摄影", "策划", "social", "volunteer", "组织协调者", "提前规划", "认真负责", "丰富经历"} {
		if !containsString(profile.Tags, want) {
			t.Fatalf("tags=%v, missing %q", profile.Tags, want)
		}
	}
	if profile.Scores.TimeScore != 70 || profile.Scores.GoalScore != 80 || profile.Scores.CommunicationScore != 75 {
		t.Fatalf("scores should reflect expanded dimensions: %#v", profile.Scores)
	}
	if !strings.Contains(profile.Summary, "摄影") || !strings.Contains(profile.Summary, "参加校园活动") {
		t.Fatalf("summary does not describe expanded profile: %s", profile.Summary)
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

func containsString(values StringList, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
