package service

import "strings"

func destinationDisplayName(value string) string {
	trimmed := strings.TrimSpace(value)
	normalized := strings.ToLower(trimmed)
	switch normalized {
	case "":
		return "未登记"
	case "observation":
		return "留观"
	case "return_class", "back_to_class", "classroom":
		return "返回班级"
	case "urgent":
		return "紧急处理"
	case "hospital":
		return "转诊"
	case "referred":
		return "转外院"
	case "leave_school":
		return "离校就医"
	case "back_to_dorm", "dormitory":
		return "返回宿舍"
	case "home":
		return "离校回家"
	case "unknown":
		return "未登记"
	default:
		if looksLikeCode(trimmed) {
			return "其他去向"
		}
		return trimmed
	}
}

func periodDisplayName(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "daily":
		return "日报"
	case "weekly":
		return "周报"
	case "monthly":
		return "月报"
	default:
		return "报表"
	}
}

func looksLikeCode(value string) bool {
	hasLetter := false
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			hasLetter = true
			continue
		}
		if (r >= '0' && r <= '9') || r == '_' || r == '-' || r == ' ' {
			continue
		}
		return false
	}
	return hasLetter
}
