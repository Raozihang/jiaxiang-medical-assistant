package service

import "testing"

func TestDestinationDisplayNameLocalizesKnownCodes(t *testing.T) {
	cases := map[string]string{
		"return_class":  "返回班级",
		"back_to_class": "返回班级",
		"classroom":     "返回班级",
		"hospital":      "转诊",
		"referred":      "转外院",
		"leave_school":  "离校就医",
		"back_to_dorm":  "返回宿舍",
		"observation":   "留观",
		"unknown":       "未登记",
	}

	for input, want := range cases {
		if got := destinationDisplayName(input); got != want {
			t.Fatalf("destinationDisplayName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDestinationDisplayNameKeepsChineseFreeText(t *testing.T) {
	if got := destinationDisplayName("医务室观察区"); got != "医务室观察区" {
		t.Fatalf("expected Chinese free text to be preserved, got %q", got)
	}
}

func TestPeriodDisplayName(t *testing.T) {
	if got := periodDisplayName("weekly"); got != "周报" {
		t.Fatalf("expected weekly to be localized, got %q", got)
	}
}
