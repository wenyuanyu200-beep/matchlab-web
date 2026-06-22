package match

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/questionnaire"
)

func ScoreActivity(signals UserSignals, candidate activity.Activity) Recommendation {
	interestMatches := matchingValues([]string(signals.Answers.Interests), []string(candidate.Tags))
	skillMatches := matchingValues([]string(signals.Answers.Skills), []string(candidate.PreferredTags))
	activityTypes := append([]string{}, []string(signals.Answers.ActivityTypes)...)
	activityTypes = append(activityTypes, []string(signals.Answers.PreferredActivityTypes)...)
	timeSignals := strings.Join(nonEmptyStrings(
		signals.Answers.AvailableTime,
		signals.Answers.ParticipationMode,
		signals.Answers.DurationPreference,
		signals.Answers.CampusOrLocation,
	), " ")
	candidateTiming := strings.Join(nonEmptyStrings(candidate.TimeText, candidate.LocationText, candidate.Description), " ")
	goalSignals := strings.Join(nonEmptyStrings(
		signals.Answers.Goal,
		signals.Answers.MainGoal,
		signals.Answers.CommunicationStyle,
		signals.Answers.TeamRole,
		signals.Answers.WorkRhythm,
		strings.Join([]string(signals.Answers.PartnerExpectation), " "),
		strings.Join([]string(signals.Answers.ParticipationPurpose), " "),
	), " ")
	candidateContext := strings.Join(append([]string{candidate.Type, candidate.Title, candidate.Description, candidate.TimeText, candidate.LocationText}, append([]string(candidate.Tags), []string(candidate.PreferredTags)...)...), " ")
	details := DetailScores{
		Interest: overlapScore([]string(signals.Answers.Interests), []string(candidate.Tags), 30),
		Skill:    overlapScore([]string(signals.Answers.Skills), []string(candidate.PreferredTags), 25),
		Type:     typeScore(activityTypes, candidate.Type),
		Time:     textScore(timeSignals, candidateTiming, 10),
		Goal:     textScore(goalSignals, candidateContext, 15),
	}
	return Recommendation{
		Activity:     candidate,
		Score:        details.Interest + details.Skill + details.Type + details.Time + details.Goal,
		DetailScores: details,
		Reason:       buildReason(interestMatches, skillMatches, signals.Answers, candidate, details),
	}
}

func RankActivities(signals UserSignals, candidates []activity.Activity, limit int) []Recommendation {
	ranked := make([]Recommendation, 0, len(candidates))
	for _, candidate := range candidates {
		ranked = append(ranked, ScoreActivity(signals, candidate))
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return ranked[i].Activity.CreatedAt.After(ranked[j].Activity.CreatedAt)
	})
	if limit < 0 {
		limit = 0
	}
	if limit < len(ranked) {
		ranked = ranked[:limit]
	}
	return ranked
}

func overlapScore(userValues, targetValues []string, cap int) int {
	user := normalizedSet(userValues)
	if len(user) == 0 {
		return 0
	}
	target := normalizedSet(targetValues)
	matches := 0
	for value := range user {
		if _, ok := target[value]; ok {
			matches++
		}
	}
	return int(math.Round(float64(matches) / float64(len(user)) * float64(cap)))
}

func typeScore(types []string, candidateType string) int {
	target := normalize(candidateType)
	for _, value := range types {
		if normalize(value) == target && target != "" {
			return 20
		}
	}
	return 0
}

func textScore(left, right string, cap int) int {
	leftNormalized := compact(left)
	rightNormalized := compact(right)
	if leftNormalized == "" || rightNormalized == "" {
		return 0
	}
	if strings.Contains(rightNormalized, leftNormalized) || strings.Contains(leftNormalized, rightNormalized) {
		return cap
	}
	leftKeywords := keywordSet(leftNormalized)
	rightKeywords := keywordSet(rightNormalized)
	if len(leftKeywords) == 0 {
		return 0
	}
	matches := 0
	for keyword := range leftKeywords {
		if _, ok := rightKeywords[keyword]; ok {
			matches++
		}
	}
	return int(math.Round(float64(matches) / float64(len(leftKeywords)) * float64(cap)))
}

func matchingValues(left, right []string) []string {
	target := normalizedSet(right)
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, value := range left {
		trimmed := strings.TrimSpace(value)
		key := normalize(trimmed)
		if key == "" || containsProhibitedTerm(trimmed) {
			continue
		}
		if _, ok := target[key]; !ok {
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

func normalizedSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if normalized := normalize(value); normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func compact(value string) string {
	return strings.ToLower(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, value))
}

func keywordSet(value string) map[string]struct{} {
	result := make(map[string]struct{})
	runes := []rune(value)
	for start := 0; start < len(runes); {
		if unicode.Is(unicode.Han, runes[start]) {
			end := start
			for end < len(runes) && unicode.Is(unicode.Han, runes[end]) {
				end++
			}
			segment := runes[start:end]
			if len(segment) == 1 {
				result[string(segment)] = struct{}{}
			} else {
				for i := 0; i < len(segment)-1; i++ {
					result[string(segment[i:i+2])] = struct{}{}
				}
			}
			start = end
			continue
		}
		end := start
		for end < len(runes) && !unicode.Is(unicode.Han, runes[end]) {
			end++
		}
		if token := strings.TrimSpace(string(runes[start:end])); token != "" {
			result[token] = struct{}{}
		}
		start = end
	}
	return result
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func buildReason(interests, skills []string, answers questionnaire.Answers, candidate activity.Activity, details DetailScores) string {
	typeLabel := activityTypeLabel(candidate.Type)
	goal := firstNonBlank(answers.MainGoal, answers.Goal)
	style := firstNonBlank(answers.TeamRole, answers.CommunicationStyle, answers.WorkRhythm)
	switch {
	case len(interests) > 0 && len(skills) > 0 && goal != "":
		return "推荐原因：该活动与你的兴趣方向“" + strings.Join(interests, "、") + "”较一致，技能标签也覆盖了“" + strings.Join(skills, "、") + "”等需求；同时它和你的参与目标“" + goal + "”有关，适合作为" + typeLabel + "候选。"
	case len(interests) > 0 && len(skills) > 0:
		return "推荐原因：该活动与你的兴趣方向“" + strings.Join(interests, "、") + "”较一致，你的技能标签也能覆盖活动需要的部分角色，适合作为校园协作候选。"
	case len(interests) > 0 && style != "":
		return "推荐原因：该活动与“" + strings.Join(interests, "、") + "”等兴趣相关，也比较适合你偏好的“" + style + "”协作方式，可以进一步了解分工和节奏。"
	case len(interests) > 0:
		return "推荐原因：该活动与你的兴趣标签“" + strings.Join(interests, "、") + "”相符，值得优先了解活动内容和参与方式。"
	case len(skills) > 0:
		return "推荐原因：你的技能倾向与活动需要的“" + strings.Join(skills, "、") + "”相符，适合在相关项目或活动中承担明确角色。"
	case details.Type > 0 && details.Time > 0:
		return "推荐原因：该活动类型符合你的参与偏好，时间、地点或参与方式也有一定契合度，可以进一步了解活动要求。"
	case details.Type > 0:
		return "推荐原因：该活动类型符合你的参与偏好，适合作为校园活动或项目协作候选。"
	case details.Time > 0 || details.Goal > 0:
		return "推荐原因：该活动的安排、地点或内容与你填写的目标存在一定契合，建议进一步了解是否适合投入。"
	default:
		return "推荐原因：该活动属于校园活动与协作场景，可以结合具体内容、时间和招募角色进一步判断是否适合。"
	}
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
		return "校园协作"
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
