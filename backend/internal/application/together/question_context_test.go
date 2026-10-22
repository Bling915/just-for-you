package together

import (
	"strings"
	"testing"
	"time"
)

func TestQuestionContextExcludesPrivateAndPreciseMediaLocation(t *testing.T) {
	lat, lng := 31.2, 121.5
	values := []ContextMoment{{Text: "共同散步", SpaceType: "shared", Visibility: "couple", Place: "精确地址", Latitude: &lat, Longitude: &lng, PhotoKeys: []string{"secret/object.jpg"}}, {Text: "我的秘密", SpaceType: "personal", Visibility: "private"}}
	context := (QuestionContextBuilder{}).Build(9, time.Now(), values)
	joined := strings.Join(context.Themes, "|")
	for _, secret := range []string{"我的秘密", "secret/object.jpg", "31.2", "121.5", "精确地址"} {
		if strings.Contains(joined, secret) {
			t.Fatalf("context leaked %q: %s", secret, joined)
		}
	}
	if !strings.Contains(joined, "共同散步") {
		t.Fatalf("missing shared theme: %s", joined)
	}
}
