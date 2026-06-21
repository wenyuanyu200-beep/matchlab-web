package match

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"matchlab/backend/internal/activity"
)

func ScoreActivity(signals UserSignals, candidate activity.Activity) Recommendation {
	interestMatches := matchingValues([]string(signals.Answers.Interests), []string(candidate.Tags))
	skillMatches := matchingValues([]string(signals.Answers.Skills), []string(candidate.PreferredTags))
	details := DetailScores{
		Interest: overlapScore([]string(signals.Answers.Interests), []string(candidate.Tags), 30),
		Skill:    overlapScore([]string(signals.Answers.Skills), []string(candidate.PreferredTags), 25),
		Type:     typeScore([]string(signals.Answers.ActivityTypes), candidate.Type),
		Time:     textScore(signals.Answers.AvailableTime, candidate.TimeText, 10),
		Goal:     textScore(signals.Answers.Goal, strings.Join(append([]string{candidate.Title, candidate.Description}, []string(candidate.Tags)...), " "), 15),
	}
	return Recommendation{
		Activity:     candidate,
		Score:        details.Interest + details.Skill + details.Type + details.Time + details.Goal,
		DetailScores: details,
		Reason:       buildReason(interestMatches, skillMatches, details),
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
	for _, term := range []string{"交友", "脱单", "约会"} {
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

func buildReason(interests, skills []string, details DetailScores) string {
	switch {
	case len(interests) > 0 && len(skills) > 0:
		return "推荐原因：你的兴趣标签与该活动的‘" + strings.Join(interests, "、") + "’高度重合，同时你的技能倾向与‘" + strings.Join(skills, "、") + "’相符，适合参与该类校园竞赛或项目协作。"
	case len(interests) > 0:
		return "推荐原因：你的兴趣标签与该活动的‘" + strings.Join(interests, "、") + "’相符，值得优先了解该校园活动。"
	case len(skills) > 0:
		return "推荐原因：你的技能倾向与活动需要的‘" + strings.Join(skills, "、") + "’相符，适合参与相关项目协作。"
	case details.Type > 0 && details.Time > 0:
		return "推荐原因：该活动类型符合你的参与偏好，时间安排也较为契合，可进一步了解活动内容。"
	case details.Type > 0:
		return "推荐原因：该活动类型符合你的参与偏好，适合作为校园活动或项目协作候选。"
	case details.Time > 0 || details.Goal > 0:
		return "推荐原因：该活动的安排或内容与你填写的目标存在一定契合，建议进一步了解。"
	default:
		return "推荐原因：该活动属于校园活动与项目协作场景，可结合具体内容进一步判断是否适合。"
	}
}
