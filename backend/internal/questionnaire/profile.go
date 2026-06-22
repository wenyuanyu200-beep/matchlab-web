package questionnaire

import (
	"fmt"
	"strings"
)

const missingDimensionScore = 20

func GenerateProfile(mode string, answers Answers) GeneratedProfile {
	answers = normalizeAnswers(answers)
	activityTypes := answers.PreferredActivityTypes
	if len(activityTypes) == 0 {
		activityTypes = answers.ActivityTypes
	}
	goal := firstNonBlank(answers.MainGoal, answers.Goal)

	return GeneratedProfile{
		ProfileType: strings.TrimSpace(mode),
		Tags:        profileTags(answers, activityTypes),
		Scores: ProfileScores{
			InterestScore:      presentScore(len(answers.Interests) > 0, 80),
			SkillScore:         presentScore(len(answers.Skills) > 0, 75),
			TimeScore:          presentScore(hasAnyText(answers.AvailableTime, answers.ParticipationMode, answers.CampusOrLocation), 70),
			GoalScore:          presentScore(goal != "", 80),
			CommunicationScore: presentScore(hasAnyText(answers.CommunicationStyle, answers.TeamRole, answers.WorkRhythm), 75),
		},
		Summary: buildSummary(answers),
	}
}

func normalizeAnswers(answers Answers) Answers {
	answers.Interests = normalizeList(answers.Interests)
	answers.Skills = normalizeList(answers.Skills)
	answers.ActivityTypes = normalizeList(answers.ActivityTypes)
	answers.PreferredActivityTypes = normalizeList(answers.PreferredActivityTypes)
	answers.Experiences = normalizeList(answers.Experiences)
	answers.PartnerExpectation = normalizeList(answers.PartnerExpectation)
	answers.AvoidPoints = normalizeList(answers.AvoidPoints)
	answers.ParticipationPurpose = normalizeList(answers.ParticipationPurpose)
	answers.Hobbies = strings.TrimSpace(answers.Hobbies)
	answers.ExploreFields = strings.TrimSpace(answers.ExploreFields)
	answers.AvailableTime = strings.TrimSpace(answers.AvailableTime)
	answers.Goal = strings.TrimSpace(answers.Goal)
	answers.MainGoal = strings.TrimSpace(answers.MainGoal)
	answers.SkillLevel = strings.TrimSpace(answers.SkillLevel)
	answers.MBTI = strings.TrimSpace(answers.MBTI)
	answers.CommunicationStyle = strings.TrimSpace(answers.CommunicationStyle)
	answers.TeamRole = strings.TrimSpace(answers.TeamRole)
	answers.WorkRhythm = strings.TrimSpace(answers.WorkRhythm)
	answers.ParticipationMode = strings.TrimSpace(answers.ParticipationMode)
	answers.DurationPreference = strings.TrimSpace(answers.DurationPreference)
	answers.CampusOrLocation = strings.TrimSpace(answers.CampusOrLocation)
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
	for _, term := range []string{"交友", "脱单", "约会", "陌生人社交", "算命", "占卜", "恋爱匹配"} {
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

func hasAnyText(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func profileTags(answers Answers, activityTypes StringList) StringList {
	styleTags := StringList{}
	for _, value := range []string{
		answers.SkillLevel,
		answers.CommunicationStyle,
		answers.TeamRole,
		answers.WorkRhythm,
		answers.ParticipationMode,
		answers.DurationPreference,
		firstNonBlank(answers.MainGoal, answers.Goal),
	} {
		if strings.TrimSpace(value) != "" {
			styleTags = append(styleTags, value)
		}
	}
	styleTags = append(styleTags, answers.PartnerExpectation...)
	styleTags = append(styleTags, answers.ParticipationPurpose...)
	return mergeUnique(answers.Interests, answers.Skills, activityTypes, answers.Experiences, styleTags)
}

func buildSummary(answers Answers) string {
	interests := take(answers.Interests, 3)
	skills := take(answers.Skills, 2)
	activityTypes := answers.PreferredActivityTypes
	if len(activityTypes) == 0 {
		activityTypes = answers.ActivityTypes
	}
	activityFocus := activityTypeSummary(activityTypes)
	goal := firstNonBlank(answers.MainGoal, answers.Goal)
	style := firstNonBlank(answers.TeamRole, answers.CommunicationStyle, answers.WorkRhythm)

	switch {
	case len(interests) > 0 && len(skills) > 0 && goal != "":
		return fmt.Sprintf("你关注%s，具备%s等能力，当前目标是%s。MatchLab 会优先推荐%s中与兴趣、技能、时间和协作方式更契合的机会。", strings.Join(interests, "、"), strings.Join(skills, "、"), goal, activityFocus)
	case len(interests) > 0 && len(skills) > 0:
		return fmt.Sprintf("你关注%s，具备%s等能力，适合从%s中寻找目标清晰、分工明确的校园协作机会。", strings.Join(interests, "、"), strings.Join(skills, "、"), activityFocus)
	case len(interests) > 0 && style != "":
		return fmt.Sprintf("你关注%s，协作上偏向%s，适合参与%s，也适合与节奏稳定、沟通清晰的伙伴一起推进。", strings.Join(interests, "、"), style, activityFocus)
	case len(interests) > 0:
		return fmt.Sprintf("你关注%s，适合优先了解%s中的相关活动，再结合时间和目标选择合适的协作机会。", strings.Join(interests, "、"), activityFocus)
	case len(skills) > 0:
		return fmt.Sprintf("你具备%s等能力，适合在%s中承担清晰角色，并通过实际协作继续完善个人画像。", strings.Join(skills, "、"), activityFocus)
	default:
		return "你偏向校园活动与项目协作场景，可以通过补充兴趣、技能、时间和协作方式，让推荐更接近真实需求。"
	}
}

func activityTypeSummary(types StringList) string {
	labels := make([]string, 0, len(types))
	for _, value := range types {
		labels = append(labels, activityTypeLabel(value))
	}
	if len(labels) == 0 {
		return "校园协作"
	}
	return strings.Join(labels, "和")
}

func activityTypeLabel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "competition":
		return "比赛组队"
	case "project":
		return "项目合作"
	case "study":
		return "学习搭子"
	case "club":
		return "社团活动"
	case "volunteer":
		return "志愿活动"
	case "workshop":
		return "讲座沙龙"
	case "social":
		return "兴趣活动"
	case "startup":
		return "创业招募"
	case "parttime":
		return "短期协作"
	default:
		return strings.TrimSpace(value)
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func take(values StringList, limit int) []string {
	if len(values) <= limit {
		return []string(values)
	}
	return []string(values[:limit])
}
