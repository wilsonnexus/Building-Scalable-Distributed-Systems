package classifier

import (
	"math"
	"strings"
	"time"
)

func Classify(logText string) (string, string, float64) {
	text := strings.ToLower(logText)

	switch {
	case strings.Contains(text, "timeout"):
		return "timeout", "Check service latency, retry settings, and downstream availability.", 0.82
	case strings.Contains(text, "connection refused") || strings.Contains(text, "dns") || strings.Contains(text, "network"):
		return "infrastructure_issue", "Check networking, DNS, service discovery, and container health.", 0.78
	case strings.Contains(text, "dependency") || strings.Contains(text, "package") || strings.Contains(text, "artifact"):
		return "dependency_failure", "Verify dependency versions, artifact availability, and external service health.", 0.74
	case strings.Contains(text, "flaky") || strings.Contains(text, "intermittent"):
		return "flaky_test", "Rerun the job, isolate unstable tests, and review timing assumptions.", 0.67
	default:
		base := 0.55 + 0.1*math.Sin(float64(time.Now().UnixNano()%1000))
		if base < 0.5 {
			base = 0.5
		}
		if base > 0.75 {
			base = 0.75
		}
		return "unknown_failure", "Review recent logs and route for manual triage or fallback handling.", base
	}
}
