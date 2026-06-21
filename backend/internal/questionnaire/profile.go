package questionnaire

import (
	"fmt"
	"strings"
)

const missingDimensionScore = 20

func GenerateProfile(mode string, answers Answers) GeneratedProfile {
	answers.Interests = normalizeList(answers.Interests)
	answers.Skills = normalizeList(answers.Skills)
	answers.ActivityTypes = normalizeList(answers.ActivityTypes)

	return GeneratedProfile{
		ProfileType: strings.TrimSpace(mode),
		Tags:        mergeUnique(answers.Interests, answers.Skills, answers.ActivityTypes),
		Scores: ProfileScores{
			InterestScore:      presentScore(len(answers.Interests) > 0, 80),
			SkillScore:         presentScore(len(answers.Skills) > 0, 75),
			TimeScore:          presentScore(strings.TrimSpace(answers.AvailableTime) != "", 70),
			GoalScore:          presentScore(strings.TrimSpace(answers.Goal) != "", 80),
			CommunicationScore: presentScore(strings.TrimSpace(answers.CommunicationStyle) != "", 75),
		},
		Summary: buildSummary(answers),
	}
}

func normalizeAnswers(answers Answers) Answers {
	answers.Interests = normalizeList(answers.Interests)
	answers.Skills = normalizeList(answers.Skills)
	answers.ActivityTypes = normalizeList(answers.ActivityTypes)
	answers.AvailableTime = strings.TrimSpace(answers.AvailableTime)
	answers.Goal = strings.TrimSpace(answers.Goal)
	answers.CommunicationStyle = strings.TrimSpace(answers.CommunicationStyle)
	return answers
}

func normalizeList(values []string) StringList {
	result := make(StringList, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		key := strings.ToLower(trimmed)
		if trimmed == "" || containsProhibitedTerm(trimmed) {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func containsProhibitedTerm(value string) bool {
	for _, term := range []string{"交友", "脱单", "约会"} {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func mergeUnique(groups ...StringList) StringList {
	merged := make(StringList, 0)
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return normalizeList(merged)
}

func presentScore(present bool, score int) int {
	if present {
		return score
	}
	return missingDimensionScore
}

func buildSummary(answers Answers) string {
	interests := take(answers.Interests, 3)
	skills := take(answers.Skills, 2)
	activityFocus := activityTypeSummary(answers.ActivityTypes)

	switch {
	case len(interests) > 0 && len(skills) > 0:
		return fmt.Sprintf("该用户偏向%s，关注%s，具备%s相关兴趣，适合参与校园活动与项目协作。", activityFocus, strings.Join(interests, "、"), strings.Join(skills, "、"))
	case len(interests) > 0:
		return fmt.Sprintf("该用户偏向%s，关注%s，适合参与相关校园活动与项目协作。", activityFocus, strings.Join(interests, "、"))
	case len(skills) > 0:
		return fmt.Sprintf("该用户偏向%s，具备%s相关兴趣，适合参与校园项目协作。", activityFocus, strings.Join(skills, "、"))
	default:
		return "该用户偏向校园活动与项目协作，可通过参与活动进一步完善能力标签。"
	}
}

func activityTypeSummary(types StringList) string {
	labels := make([]string, 0, len(types))
	for _, value := range types {
		switch strings.ToLower(value) {
		case "competition":
			labels = append(labels, "竞赛组队")
		case "project":
			labels = append(labels, "项目协作")
		default:
			labels = append(labels, value)
		}
	}
	if len(labels) == 0 {
		return "校园协作"
	}
	return strings.Join(labels, "和")
}

func take(values StringList, limit int) []string {
	if len(values) <= limit {
		return []string(values)
	}
	return []string(values[:limit])
}
